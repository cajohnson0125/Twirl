package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/cajohnson0125/Twirl/internal/agent"
	"github.com/cajohnson0125/Twirl/internal/pubsub"
	"github.com/cajohnson0125/Twirl/internal/state"
	"github.com/cajohnson0125/Twirl/internal/workflow"
)

// --- Linear path ---

func TestLinearPath(t *testing.T) {
	g := workflow.NewGraph("a")
	g.AddNode(&workflow.Node{ID: "a", Role: "brainstorm"})
	g.AddNode(&workflow.Node{ID: "b", Role: "research"})
	g.AddNode(&workflow.Node{ID: "c", Role: "scribe"})
	g.AddEdge("a", "b", nil)
	g.AddEdge("b", "c", nil)

	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm,
			&state.Result{Status: state.StatusCompleted})
	})
	regs.Register(agent.Research, func() agent.Agent {
		return agent.NewStubAgent(agent.Research,
			&state.Result{Status: state.StatusCompleted})
	})
	regs.Register(agent.Scribe, func() agent.Agent {
		return agent.NewStubAgent(agent.Scribe,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	hitlIn := make(chan state.HITLResponse, 1)
	store := state.NewStore(t.TempDir())
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	err := e.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	s := e.State()
	if s.Status != state.StatusCompleted {
		t.Errorf("Status: got %d, want Completed", s.Status)
	}
	if len(s.ActiveNodes) != 0 {
		t.Errorf("ActiveNodes: got %v, want empty",
			s.ActiveNodes)
	}
	if len(s.AuditLog) == 0 {
		t.Error("AuditLog: expected entries")
	}
}

// --- Conditional routing ---

func TestConditionalRouting(t *testing.T) {
	g := workflow.NewGraph("review")
	g.AddNode(&workflow.Node{ID: "review", Role: "code_review"})
	g.AddNode(&workflow.Node{ID: "triage", Role: "triage"})
	g.AddNode(&workflow.Node{ID: "scribe", Role: "scribe"})

	g.AddEdge("review", "triage",
		func(_ *state.State, r *state.Result) bool {
			return r.Severity > 0
		})
	g.AddEdge("review", "scribe",
		func(_ *state.State, r *state.Result) bool {
			return r.Severity == 0
		})
	g.AddEdge("triage", "scribe", nil)

	t.Run("issues_found", func(t *testing.T) {
		regs := agent.NewRegistry()
		regs.Register(agent.CodeReview, func() agent.Agent {
			return agent.NewStubAgent(agent.CodeReview,
				&state.Result{
					Status:   state.StatusCompleted,
					Severity: 3,
				})
		})
		regs.Register(agent.Triage, func() agent.Agent {
			return agent.NewStubAgent(agent.Triage,
				&state.Result{Status: state.StatusCompleted})
		})
		regs.Register(agent.Scribe, func() agent.Agent {
			return agent.NewStubAgent(agent.Scribe,
				&state.Result{Status: state.StatusCompleted})
		})

		bus := pubsub.NewBus(64)
		hitlIn := make(chan state.HITLResponse, 1)
		store := state.NewStore(t.TempDir())
		e := NewEngine("test", g, store, regs, bus, hitlIn)

		if err := e.Run(context.Background()); err != nil {
			t.Fatalf("Run: %v", err)
		}

		found := false
		for _, ev := range e.State().AuditLog {
			if ev.Type == "RESULT" && ev.NodeID == "triage" {
				found = true
			}
		}
		if !found {
			t.Error("triage node was not executed")
		}
	})

	t.Run("no_issues", func(t *testing.T) {
		regs := agent.NewRegistry()
		regs.Register(agent.CodeReview, func() agent.Agent {
			return agent.NewStubAgent(agent.CodeReview,
				&state.Result{
					Status:   state.StatusCompleted,
					Severity: 0,
				})
		})
		regs.Register(agent.Triage, func() agent.Agent {
			return agent.NewStubAgent(agent.Triage,
				&state.Result{Status: state.StatusCompleted})
		})
		regs.Register(agent.Scribe, func() agent.Agent {
			return agent.NewStubAgent(agent.Scribe,
				&state.Result{Status: state.StatusCompleted})
		})

		bus := pubsub.NewBus(64)
		hitlIn := make(chan state.HITLResponse, 1)
		store := state.NewStore(t.TempDir())
		e := NewEngine("test", g, store, regs, bus, hitlIn)

		if err := e.Run(context.Background()); err != nil {
			t.Fatalf("Run: %v", err)
		}

		for _, ev := range e.State().AuditLog {
			if ev.NodeID == "triage" {
				t.Error("triage should not execute " +
					"when no issues found")
			}
		}
	})
}

// --- Loop-back ---

func TestLoopBack(t *testing.T) {
	g := workflow.NewGraph("plan")
	g.AddNode(&workflow.Node{ID: "plan", Role: "plan"})
	g.AddNode(&workflow.Node{ID: "review", Role: "plan_review"})
	g.AddNode(&workflow.Node{ID: "done", Role: "scribe"})

	g.AddEdge("plan", "review", nil)
	g.AddEdge("review", "done",
		func(_ *state.State, r *state.Result) bool {
			return r.Action == "PASS"
		})
	// Loop-back: review fails -> back to plan.
	g.AddEdge("review", "plan",
		func(_ *state.State, r *state.Result) bool {
			return r.Action == "FAIL"
		})
	// "done" has no outgoing edges = terminal.

	passCount := 0
	regs := agent.NewRegistry()
	regs.Register(agent.Plan, func() agent.Agent {
		return agent.NewStubAgent(agent.Plan,
			&state.Result{Status: state.StatusCompleted})
	})
	regs.Register(agent.PlanReview, func() agent.Agent {
		passCount++
		if passCount < 3 {
			return agent.NewStubAgent(agent.PlanReview,
				&state.Result{
					Status: state.StatusCompleted,
					Action: "FAIL",
				})
		}
		return agent.NewStubAgent(agent.PlanReview,
			&state.Result{
				Status: state.StatusCompleted,
				Action: "PASS",
			})
	})
	regs.Register(agent.Scribe, func() agent.Agent {
		return agent.NewStubAgent(agent.Scribe,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	hitlIn := make(chan state.HITLResponse, 1)
	store := state.NewStore(t.TempDir())
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	if err := e.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	planCount := 0
	for _, ev := range e.State().AuditLog {
		if ev.NodeID == "plan" && ev.Type == "RESULT" {
			planCount++
		}
	}
	if planCount != 3 {
		t.Errorf("plan executed %d times, want 3", planCount)
	}
}

// --- Parallel execution ---

func TestParallelExecution(t *testing.T) {
	g := workflow.NewGraph("fork")
	g.AddNode(&workflow.Node{ID: "fork", Role: "execution"})
	g.AddNode(&workflow.Node{ID: "exec1", Role: "execution"})
	g.AddNode(&workflow.Node{ID: "exec2", Role: "execution"})
	g.AddNode(&workflow.Node{ID: "join", Role: "code_review"})

	g.AddEdge("fork", "exec1", nil)
	g.AddEdge("fork", "exec2", nil)
	g.AddEdge("exec1", "join", nil)
	g.AddEdge("exec2", "join", nil)

	regs := agent.NewRegistry()
	regs.Register(agent.Execution, func() agent.Agent {
		return agent.NewStubAgent(agent.Execution,
			&state.Result{Status: state.StatusCompleted})
	})
	regs.Register(agent.CodeReview, func() agent.Agent {
		return agent.NewStubAgent(agent.CodeReview,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	hitlIn := make(chan state.HITLResponse, 1)
	store := state.NewStore(t.TempDir())
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	// Start with two active nodes to test parallel.
	e.state.ActiveNodes = []string{"exec1", "exec2"}

	if err := e.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	execCount := 0
	for _, ev := range e.State().AuditLog {
		if ev.Type == "RESULT" &&
			(ev.NodeID == "exec1" || ev.NodeID == "exec2") {
			execCount++
		}
	}
	if execCount != 2 {
		t.Errorf("exec nodes ran %d times, want 2", execCount)
	}

	found := false
	for _, ev := range e.State().AuditLog {
		if ev.NodeID == "join" {
			found = true
		}
	}
	if !found {
		t.Error("join node was not executed")
	}
}

// --- HITL gate ---

func TestHITLGate(t *testing.T) {
	g := workflow.NewGraph("gate")
	g.AddNode(&workflow.Node{
		ID:   "gate",
		Role: "hitl_gate",
		Execute: func(_ context.Context, _ *state.State) (*state.Result, error) {
			return &state.Result{
				Status: state.StatusPausedHITL,
				HITLRequest: &state.HITLRequest{
					ID:      "h1",
					Prompt:  "Proceed?",
					Options: []string{"yes", "no"},
				},
			}, nil
		},
	})
	g.AddNode(&workflow.Node{ID: "done", Role: "scribe"})
	g.AddEdge("gate", "done", nil)

	regs := agent.NewRegistry()
	regs.Register(agent.Scribe, func() agent.Agent {
		return agent.NewStubAgent(agent.Scribe,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	gateCh := bus.Subscribe(pubsub.EventGate)
	hitlIn := make(chan state.HITLResponse, 1)
	store := state.NewStore(t.TempDir())
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	done := make(chan error, 1)
	go func() {
		done <- e.Run(context.Background())
	}()

	// Wait for the gate event, then send response.
	select {
	case ev, ok := <-gateCh:
		if !ok {
			t.Fatal("gate channel closed unexpectedly")
		}
		if ev.Prompt != "Proceed?" {
			t.Errorf("prompt: got %q, want %q",
				ev.Prompt, "Proceed?")
		}
		hitlIn <- state.HITLResponse{
			ID:     "h1",
			Choice: "yes",
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for gate event")
	}

	if err := <-done; err != nil {
		t.Fatalf("Run: %v", err)
	}
	s := e.State()
	if s.Status != state.StatusCompleted {
		t.Errorf("Status: got %d, want Completed", s.Status)
	}
	if s.Context["hitl_choice"] != "yes" {
		t.Errorf("hitl_choice: got %q, want %q",
			s.Context["hitl_choice"], "yes")
	}
}

// --- Cancel ---

func TestCancel(t *testing.T) {
	g := workflow.NewGraph("start")
	g.AddNode(&workflow.Node{ID: "start", Role: "brainstorm"})
	// Node with one outgoing edge back to itself. Cancel will
	// stop the loop via context.
	g.AddEdge("start", "start", nil)

	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	doneCh := bus.Subscribe(pubsub.EventAgentDone)
	hitlIn := make(chan state.HITLResponse, 1)
	store := state.NewStore(t.TempDir())
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- e.Run(ctx)
	}()

	// Wait for at least one iteration, then cancel.
	select {
	case <-doneCh:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for first node")
	}

	err := <-done
	if err == nil {
		t.Error("expected error on cancel")
	}
	if e.State().Status != state.StatusFailed {
		t.Errorf("Status: got %d, want Failed",
			e.State().Status)
	}
}

// --- Resume ---

func TestResume(t *testing.T) {
	dir := t.TempDir()

	g := workflow.NewGraph("a")
	g.AddNode(&workflow.Node{ID: "a", Role: "brainstorm"})
	g.AddNode(&workflow.Node{ID: "b", Role: "research"})
	g.AddEdge("a", "b", nil)

	// First run: execute both nodes in a loop that
	// self-edges, then cancel after "a" completes.
	regs1 := agent.NewRegistry()
	callCount := 0
	regs1.Register(agent.Brainstorm, func() agent.Agent {
		callCount++
		if callCount == 1 {
			return agent.NewStubAgent(agent.Brainstorm,
				&state.Result{Status: state.StatusCompleted})
		}
		// After the first call, block until cancel.
		return agent.NewStubAgentWithError(agent.Brainstorm,
			context.Canceled)
	})
	regs1.Register(agent.Research, func() agent.Agent {
		return agent.NewStubAgent(agent.Research,
			&state.Result{Status: state.StatusCompleted})
	})

	bus1 := pubsub.NewBus(64)
	firstDone := bus1.Subscribe(pubsub.EventAgentDone)
	hitlIn1 := make(chan state.HITLResponse, 1)
	store1 := state.NewStore(dir)

	ctx1, cancel1 := context.WithCancel(context.Background())
	e1 := NewEngine("test", g, store1, regs1, bus1, hitlIn1)

	go func() {
		e1.Run(ctx1)
	}()

	// Wait for node "a" to finish once.
	select {
	case ev := <-firstDone:
		if ev.NodeID != "a" {
			t.Fatalf("expected first done from 'a', got %q",
				ev.NodeID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for first node")
	}

	// State should now have "b" as the next active node.
	// The engine will try to execute "b" next, but we cancel
	// so the persisted state has ActiveNodes=["b"].
	cancel1()

	// Poll until state is persisted with ActiveNodes=["b"].
	deadline := time.Now().Add(2 * time.Second)
	for {
		loaded, err := store1.Load()
		if err == nil && len(loaded.ActiveNodes) > 0 &&
			loaded.ActiveNodes[0] == "b" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for state persist")
		}
		time.Sleep(time.Millisecond)
	}

	// Now resume from the saved state.
	regs2 := agent.NewRegistry()
	regs2.Register(agent.Research, func() agent.Agent {
		return agent.NewStubAgent(agent.Research,
			&state.Result{Status: state.StatusCompleted})
	})

	bus2 := pubsub.NewBus(64)
	hitlIn2 := make(chan state.HITLResponse, 1)
	store2 := state.NewStore(dir)

	e2, err := ResumeEngine(
		g, store2, regs2, bus2, hitlIn2)
	if err != nil {
		t.Fatalf("ResumeEngine: %v", err)
	}

	if len(e2.State().ActiveNodes) == 0 ||
		e2.State().ActiveNodes[0] != "b" {
		t.Errorf("ActiveNodes: got %v, want [b]",
			e2.State().ActiveNodes)
	}
	if len(e2.State().AuditLog) == 0 {
		t.Errorf("AuditLog: got %d entries, want > 0",
			len(e2.State().AuditLog))
	}

	if err := e2.Run(context.Background()); err != nil {
		t.Fatalf("Run resumed: %v", err)
	}
	if e2.State().Status != state.StatusCompleted {
		t.Errorf("Status: got %d, want Completed",
			e2.State().Status)
	}
}

// --- State persistence ---

func TestStatePersistedAfterEachNode(t *testing.T) {
	g := workflow.NewGraph("a")
	g.AddNode(&workflow.Node{ID: "a", Role: "brainstorm"})
	g.AddNode(&workflow.Node{ID: "b", Role: "research"})
	g.AddEdge("a", "b", nil)

	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm,
			&state.Result{Status: state.StatusCompleted})
	})
	regs.Register(agent.Research, func() agent.Agent {
		return agent.NewStubAgent(agent.Research,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	hitlIn := make(chan state.HITLResponse, 1)
	dir := t.TempDir()
	store := state.NewStore(dir)
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	if err := e.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify final state was persisted.
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Status != state.StatusCompleted {
		t.Errorf("persisted status: got %d, want Completed",
			loaded.Status)
	}
}

// --- Event bus integration ---

func TestPublishesAgentEvents(t *testing.T) {
	g := workflow.NewGraph("a")
	g.AddNode(&workflow.Node{ID: "a", Role: "brainstorm"})

	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm,
			&state.Result{Status: state.StatusCompleted})
	})

	bus := pubsub.NewBus(64)
	started := bus.Subscribe(pubsub.EventAgentStarted)
	done := bus.Subscribe(pubsub.EventAgentDone)
	hitlIn := make(chan state.HITLResponse, 1)
	store := state.NewStore(t.TempDir())
	e := NewEngine("test", g, store, regs, bus, hitlIn)

	if err := e.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	ev, ok := <-started
	if !ok || ev.Role != "brainstorm" {
		t.Errorf("started event: %+v, ok=%v", ev, ok)
	}
	ev, ok = <-done
	if !ok || ev.Role != "brainstorm" {
		t.Errorf("done event: %+v, ok=%v", ev, ok)
	}
}
