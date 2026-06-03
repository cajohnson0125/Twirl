package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/cajohnson0125/Twirl/internal/agent"
)

// Gate decides whether a specialist should be dispatched.
type Gate interface {
	Check(ctx context.Context, role agent.Role, state map[string]string) (bool, string)
}

// Router decides which specialist to dispatch based on user input.
type Router interface {
	Route(ctx context.Context, input string, state map[string]string) (agent.Role, error)
}

// Coordinator is the outer loop. It receives user input, decides
// what specialist to dispatch, runs the coordination layer cycle,
// and sends responses back to the user.
type Coordinator struct {
	userIn  <-chan string
	userOut chan<- string
	regs    *agent.Registry
	gate    Gate
	router  Router
	store   *Store
	bus     *Bus
	ctx     context.Context
	cancel  context.CancelFunc

	mu      sync.Mutex
	context map[string]string
	audit   []AuditEntry
}

// NewCoordinator creates a coordinator wired to the given channels.
// Gate and Router default to always-approve stubs if nil.
func NewCoordinator(
	userIn <-chan string,
	userOut chan<- string,
	regs *agent.Registry,
	opts ...CoordinatorOpt,
) *Coordinator {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Coordinator{
		userIn:  userIn,
		userOut: userOut,
		regs:    regs,
		gate:    &stubGate{},
		router:  &stubRouter{},
		ctx:     ctx,
		cancel:  cancel,
		context: make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CoordinatorOpt configures a coordinator.
type CoordinatorOpt func(*Coordinator)

// WithGate sets the gate implementation.
func WithGate(g Gate) CoordinatorOpt {
	return func(c *Coordinator) { c.gate = g }
}

// WithRouter sets the router implementation.
func WithRouter(r Router) CoordinatorOpt {
	return func(c *Coordinator) { c.router = r }
}

// WithBus sets the event bus for publishing.
func WithBus(b *Bus) CoordinatorOpt {
	return func(c *Coordinator) { c.bus = b }
}

// WithStore sets the state store for persistence.
func WithStore(s *Store) CoordinatorOpt {
	return func(c *Coordinator) { c.store = s }
}

// Cancel stops the coordinator.
func (c *Coordinator) Cancel() { c.cancel() }

// Context returns a copy of the current project context.
func (c *Coordinator) Context() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make(map[string]string, len(c.context))
	for k, v := range c.context {
		cp[k] = v
	}
	return cp
}

// Run drives the main loop. It blocks until the context is
// cancelled or the userIn channel is closed.
func (c *Coordinator) Run() error {
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case input, ok := <-c.userIn:
			if !ok {
				return nil
			}
			result, err := c.coordinate(input)
			if err != nil {
				c.send(fmt.Sprintf("Error: %s", err))
				continue
			}
			c.sendResult(result)
		}
	}
}

// coordinate runs one coordination cycle: route, gate, dispatch,
// collect. It returns the specialist result or an error.
func (c *Coordinator) coordinate(input string) (*agent.Result, error) {
	role := c.route(input)
	approved, reason := c.gateCheck(role)
	if !approved {
		return nil, fmt.Errorf("gate rejected %s: %s", role, reason)
	}
	result, err := c.dispatch(role, input)
	if err != nil {
		return nil, err
	}
	c.collect(result)
	return result, nil
}

// route determines which specialist to dispatch.
func (c *Coordinator) route(input string) agent.Role {
	role, err := c.router.Route(c.ctx, input, c.copyContext())
	if err != nil {
		return agent.Brainstorm
	}
	return role
}

// gate checks whether a specialist should run.
func (c *Coordinator) gateCheck(role agent.Role) (bool, string) {
	return c.gate.Check(c.ctx, role, c.copyContext())
}

// dispatch looks up the agent, builds a task, and executes it.
// It proxies HITL requests from the specialist to the user and
// forwards responses back.
func (c *Coordinator) dispatch(
	role agent.Role,
	input string,
) (*agent.Result, error) {
	a, err := c.regs.Get(role)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	ctx := make(map[string]string, len(c.context))
	for k, v := range c.context {
		ctx[k] = v
	}
	c.mu.Unlock()

	hitlPrompt := make(chan string, 4)
	hitlResp := make(chan string, 4)

	task := &agent.Task{
		Instruction:  input,
		Context:      ctx,
		HITLPrompt:   hitlPrompt,
		HITLResponse: hitlResp,
		Stream: func(token string) {
			c.send(token)
		},
	}

	c.publish(Event{Type: EventAgentStarted, Role: string(role)})

	type execResult struct {
		result *agent.Result
		err    error
	}
	done := make(chan execResult, 1)
	go func() {
		r, err := a.Execute(c.ctx, task)
		done <- execResult{r, err}
		close(hitlPrompt)
		c.publish(Event{Type: EventAgentDone, Role: string(role)})
	}()

	// Proxy HITL: forward specialist prompts to user, forward
	// user responses back to specialist.
	for {
		select {
		case prompt, ok := <-hitlPrompt:
			if !ok {
				// Specialist is done asking questions.
				r := <-done
				return r.result, r.err
			}
			c.send(prompt)
			select {
			case resp, ok := <-c.userIn:
				if ok {
					hitlResp <- resp
				}
			case <-c.ctx.Done():
				return nil, c.ctx.Err()
			}
		case <-c.ctx.Done():
			return nil, c.ctx.Err()
		}
	}
}

// collect merges the result context back into project state,
// appends an audit entry, and persists to disk.
func (c *Coordinator) collect(result *agent.Result) {
	c.mu.Lock()
	if result.Context != nil {
		for k, v := range result.Context {
			c.context[k] = v
		}
	}
	c.audit = append(c.audit, AuditEntry{
		Phase:   "collect",
		Message: result.OutputPath,
	})
	c.persistLocked()
	c.mu.Unlock()
}

// persistLocked writes state to disk. Caller must hold c.mu.
func (c *Coordinator) persistLocked() {
	if c.store == nil {
		return
	}
	st := &State{
		Context:  c.context,
		AuditLog: c.audit,
	}
	_ = c.store.Save(st)
}

func (c *Coordinator) copyContext() map[string]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make(map[string]string, len(c.context))
	for k, v := range c.context {
		cp[k] = v
	}
	return cp
}

func (c *Coordinator) send(msg string) {
	select {
	case c.userOut <- msg:
	case <-c.ctx.Done():
	}
}

func (c *Coordinator) publish(e Event) {
	if c.bus != nil {
		c.bus.Publish(e)
	}
}

func (c *Coordinator) sendResult(result *agent.Result) {
	msg := "Done"
	if result.OutputPath != "" {
		msg = fmt.Sprintf("Done: %s", result.OutputPath)
	}
	c.send(msg)
}

// stubGate always approves.
type stubGate struct{}

func (s *stubGate) Check(
	_ context.Context, _ agent.Role, _ map[string]string,
) (bool, string) {
	return true, ""
}

// stubRouter always returns Brainstorm.
type stubRouter struct{}

func (s *stubRouter) Route(
	_ context.Context, _ string, _ map[string]string,
) (agent.Role, error) {
	return agent.Brainstorm, nil
}
