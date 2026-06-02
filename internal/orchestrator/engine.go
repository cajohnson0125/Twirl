package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/cajohnson0125/Twirl/internal/agent"
	"github.com/cajohnson0125/Twirl/internal/pubsub"
	"github.com/cajohnson0125/Twirl/internal/state"
	"github.com/cajohnson0125/Twirl/internal/workflow"
)

// Engine is the coordinator that walks the workflow graph,
// dispatches agents, handles HITL gates, and persists state.
// It is the single master goroutine of the orchestration layer.
type Engine struct {
	graph  *workflow.Graph
	state  *state.State
	store  *state.Store
	regs   *agent.Registry
	bus    *pubsub.Bus
	hitlIn <-chan state.HITLResponse

	// mu protects state during concurrent node execution.
	mu sync.Mutex
}

// NewEngine creates an engine with fresh state for a new project.
func NewEngine(
	projectID string,
	graph *workflow.Graph,
	store *state.Store,
	regs *agent.Registry,
	bus *pubsub.Bus,
	hitlIn <-chan state.HITLResponse,
) *Engine {
	return &Engine{
		graph: graph,
		state: &state.State{
			ProjectID:   projectID,
			ActiveNodes: []string{string(graph.Start)},
			Status:      state.StatusRunning,
			Context:     make(map[string]string),
		},
		store:  store,
		regs:   regs,
		bus:    bus,
		hitlIn: hitlIn,
	}
}

// ResumeEngine creates an engine from persisted state. Returns
// an error if no state file exists.
func ResumeEngine(
	graph *workflow.Graph,
	store *state.Store,
	regs *agent.Registry,
	bus *pubsub.Bus,
	hitlIn <-chan state.HITLResponse,
) (*Engine, error) {
	st, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("resume: %w", err)
	}
	return &Engine{
		graph:  graph,
		state:  st,
		store:  store,
		regs:   regs,
		bus:    bus,
		hitlIn: hitlIn,
	}, nil
}

// State returns a deep copy of the current engine state.
func (e *Engine) State() *state.State {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.snapshotState()
}

// Run drives the core loop until the workflow completes, fails,
// or the context is cancelled. It persists state before every
// iteration so the workflow can be resumed after a crash.
func (e *Engine) Run(ctx context.Context) error {
	for {
		// 1. Persist state (crash recovery).
		e.mu.Lock()
		if err := e.store.Save(e.state); err != nil {
			e.mu.Unlock()
			return fmt.Errorf("persist: %w", err)
		}
		e.mu.Unlock()

		// 2. Check for cancellation.
		if err := ctx.Err(); err != nil {
			e.mu.Lock()
			e.state.Status = state.StatusFailed
			e.store.Save(e.state)
			e.mu.Unlock()
			return err
		}

		// 3. Check for completion.
		e.mu.Lock()
		active := e.state.ActiveNodes
		status := e.state.Status
		e.mu.Unlock()

		if len(active) == 0 || status == state.StatusCompleted {
			e.mu.Lock()
			e.state.Status = state.StatusCompleted
			e.store.Save(e.state)
			e.mu.Unlock()
			return nil
		}

		// 4. Handle HITL gate.
		if status == state.StatusPausedHITL {
			if err := e.handleHITL(ctx); err != nil {
				return err
			}
			continue
		}

		// 5. Execute active nodes (possibly in parallel).
		results, err := e.executeNodes(ctx, active)
		if err != nil {
			e.mu.Lock()
			e.state.Status = state.StatusFailed
			e.appendLog("ERROR", "", err.Error())
			e.store.Save(e.state)
			e.mu.Unlock()
			return err
		}

		// 6. Evaluate routing to determine next nodes.
		e.mu.Lock()
		nextNodes := e.evaluateRouting(active, results)

		// Check if any result paused for HITL.
		for _, r := range results {
			if r.Status == state.StatusPausedHITL &&
				r.HITLRequest != nil {
				e.state.Status = state.StatusPausedHITL
				e.state.PendingHITL = r.HITLRequest
				break
			}
		}

		// Always advance to next nodes. If paused for HITL,
		// the next iteration will handle the gate before
		// executing them.
		e.state.ActiveNodes = nextNodes
		e.mu.Unlock()
	}
}

// executeNodes runs all active nodes. If there are multiple,
// they execute concurrently via errgroup. Each node's Execute
// function is called — which dispatches the agent via the
// registry.
func (e *Engine) executeNodes(
	ctx context.Context,
	nodeIDs []string,
) (map[string]*state.Result, error) {
	results := make(map[string]*state.Result)
	var mu sync.Mutex

	g, gCtx := errgroup.WithContext(ctx)

	for _, id := range nodeIDs {
		id := id
		g.Go(func() error {
			node, ok := e.graph.Nodes[workflow.NodeID(id)]
			if !ok {
				return fmt.Errorf(
					"node %q not found in graph", id)
			}

			// Log dispatch.
			e.mu.Lock()
			e.appendLog("DISPATCH", id, node.Role)
			e.mu.Unlock()

			// Publish agent started event.
			e.bus.Publish(pubsub.Event{
				Type:   pubsub.EventAgentStarted,
				NodeID: id,
				Role:   node.Role,
			})

			e.mu.Lock()
			snapshot := e.snapshotState()
			e.mu.Unlock()

			// If the node has a custom Execute, use it.
			// Otherwise, dispatch via the agent registry.
			var res *state.Result
			var err error
			if node.Execute != nil {
				res, err = node.Execute(gCtx, snapshot)
			} else {
				res, err = e.dispatchAgent(
					gCtx, node.Role, snapshot)
			}

			if err != nil {
				e.bus.Publish(pubsub.Event{
					Type:   pubsub.EventError,
					NodeID: id,
					Err:    err.Error(),
				})
				return err
			}

			mu.Lock()
			results[id] = res
			mu.Unlock()

			// Merge result context into engine state.
			if res.Context != nil {
				e.mu.Lock()
				for k, v := range res.Context {
					e.state.Context[k] = v
				}
				e.mu.Unlock()
			}

			// Log the completion.
			e.mu.Lock()
			e.appendLog("RESULT", id,
				fmt.Sprintf("status=%d path=%s",
					res.Status, res.OutputPath))
			e.mu.Unlock()

			// Publish agent done event.
			e.bus.Publish(pubsub.Event{
				Type:   pubsub.EventAgentDone,
				NodeID: id,
				Role:   node.Role,
			})

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, err
	}
	return results, nil
}

// dispatchAgent looks up an agent by role in the registry,
// creates a Task from the current state, and executes it.
func (e *Engine) dispatchAgent(
	ctx context.Context,
	role string,
	s *state.State,
) (*state.Result, error) {
	a, err := e.regs.Get(agent.Role(role))
	if err != nil {
		return nil, err
	}
	task := &agent.Task{
		Instruction: fmt.Sprintf(
			"Execute %s workflow step", role),
		Context: s.Context,
	}
	return a.Execute(ctx, task)
}

// evaluateRouting determines the next set of active nodes by
// evaluating edge conditions from completed nodes.
func (e *Engine) evaluateRouting(
	completed []string,
	results map[string]*state.Result,
) []string {
	var next []string
	seen := make(map[string]bool)

	for _, nodeID := range completed {
		edges, ok := e.graph.Edges[workflow.NodeID(nodeID)]
		if !ok {
			// Terminal node — no outgoing edges.
			e.appendLog("ROUTE", nodeID, "terminal")
			continue
		}

		res := results[nodeID]
		for _, edge := range edges {
			if edge.Condition == nil ||
				edge.Condition(e.state, res) {
				if !seen[string(edge.To)] {
					next = append(next, string(edge.To))
					seen[string(edge.To)] = true
				}
				e.appendLog("ROUTE", nodeID,
					fmt.Sprintf("-> %s", edge.To))
			}
		}
	}

	return next
}

// handleHITL blocks until the user responds to a pending HITL
// gate. It sends the request to the TUI via the event bus and
// waits on the hitlIn channel.
func (e *Engine) handleHITL(ctx context.Context) error {
	e.mu.Lock()
	req := e.state.PendingHITL
	e.mu.Unlock()

	if req == nil {
		return errors.New(
			"HITL paused but no pending request")
	}

	// Publish gate event for TUI.
	e.bus.Publish(pubsub.Event{
		Type:    pubsub.EventGate,
		Prompt:  req.Prompt,
		Options: req.Options,
	})

	// Block until user responds or context cancelled.
	select {
	case resp := <-e.hitlIn:
		e.mu.Lock()
		// Merge user response into context.
		e.state.Context["hitl_choice"] = resp.Choice
		if resp.Input != "" {
			e.state.Context["hitl_input"] = resp.Input
		}
		e.state.PendingHITL = nil
		e.state.Status = state.StatusRunning
		e.appendLog("HITL", "",
			fmt.Sprintf("choice=%s input=%s",
				resp.Choice, resp.Input))
		e.mu.Unlock()
		return nil
	case <-ctx.Done():
		e.mu.Lock()
		e.state.Status = state.StatusFailed
		e.store.Save(e.state)
		e.mu.Unlock()
		return ctx.Err()
	}
}

// appendLog adds an entry to the audit log. Caller must hold
// e.mu.
func (e *Engine) appendLog(typ, nodeID, msg string) {
	e.state.AuditLog = append(e.state.AuditLog, state.Event{
		Type:    typ,
		NodeID:  nodeID,
		Message: msg,
	})
}

// snapshotState returns a copy of the current state for passing
// to node execute functions.
func (e *Engine) snapshotState() *state.State {
	cp := &state.State{
		ProjectID:   e.state.ProjectID,
		ActiveNodes: make([]string, len(e.state.ActiveNodes)),
		Status:      e.state.Status,
		Context:     make(map[string]string, len(e.state.Context)),
		AuditLog:    make([]state.Event, len(e.state.AuditLog)),
	}
	copy(cp.ActiveNodes, e.state.ActiveNodes)
	for k, v := range e.state.Context {
		cp.Context[k] = v
	}
	copy(cp.AuditLog, e.state.AuditLog)
	if e.state.PendingHITL != nil {
		cp.PendingHITL = &state.HITLRequest{
			ID:      e.state.PendingHITL.ID,
			Prompt:  e.state.PendingHITL.Prompt,
			Options: make([]string, len(e.state.PendingHITL.Options)),
		}
		copy(cp.PendingHITL.Options, e.state.PendingHITL.Options)
	}
	return cp
}
