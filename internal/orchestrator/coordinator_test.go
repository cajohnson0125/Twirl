package orchestrator

import (
	"context"
	"testing"

	"github.com/cajohnson0125/Twirl/internal/agent"
)

func setup() (*Coordinator, chan string, chan string) {
	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm, &agent.Result{
			OutputPath: "brainstorm.md",
			Context:    map[string]string{"approaches": "3"},
		})
	})

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs)
	return c, userIn, userOut
}

func TestCoordinator_FullLoop(t *testing.T) {
	c, userIn, userOut := setup()

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "I want to build a task manager"

	msg, ok := <-userOut
	if !ok {
		t.Fatal("userOut closed without response")
	}
	if msg != "Done: brainstorm.md" {
		t.Errorf("got %q, want %q", msg, "Done: brainstorm.md")
	}

	c.Cancel()
	err := <-done
	if err == nil {
		t.Error("expected context cancelled error")
	}
}

func TestCoordinator_ContextFlows(t *testing.T) {
	c, userIn, userOut := setup()

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test input"
	<-userOut

	ctx := c.Context()
	if ctx["approaches"] != "3" {
		t.Errorf("context approaches = %q, want %q",
			ctx["approaches"], "3")
	}

	c.Cancel()
	<-done
}

func TestCoordinator_Cancel(t *testing.T) {
	c, _, _ := setup()

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	c.Cancel()
	err := <-done
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestCoordinator_UnregisteredRole(t *testing.T) {
	regs := agent.NewRegistry()
	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs)

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"
	msg := <-userOut
	if msg == "" {
		t.Error("expected error message for unregistered role")
	}

	c.Cancel()
	<-done
}

// --- Streaming ---

func setupStreaming() (*Coordinator, chan string, chan string) {
	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStreamingStubAgent(
			agent.Brainstorm,
			&agent.Result{
				OutputPath: "brainstorm.md",
				Context:    map[string]string{"approaches": "3"},
			},
			[]string{"hello", " ", "world"},
		)
	})

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs)
	return c, userIn, userOut
}

func TestCoordinator_Streaming(t *testing.T) {
	c, userIn, userOut := setupStreaming()

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"

	var tokens []string
	for i := 0; i < 4; i++ {
		msg := <-userOut
		tokens = append(tokens, msg)
	}

	if len(tokens) != 4 {
		t.Fatalf("got %d messages, want 4", len(tokens))
	}
	if tokens[0] != "hello" {
		t.Errorf("token[0] = %q, want %q", tokens[0], "hello")
	}
	if tokens[3] != "Done: brainstorm.md" {
		t.Errorf("token[3] = %q, want done message", tokens[3])
	}

	c.Cancel()
	<-done
}

func TestCoordinator_StreamingTokensBeforeResult(t *testing.T) {
	c, userIn, userOut := setupStreaming()

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"

	first := <-userOut
	if first == "Done: brainstorm.md" {
		t.Error("done message arrived before stream tokens")
	}

	for i := 0; i < 3; i++ {
		<-userOut
	}

	c.Cancel()
	<-done
}

// --- HITL ---

func setupHITL() (*Coordinator, chan string, chan string) {
	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewHITLStubAgent(
			agent.Brainstorm,
			&agent.Result{
				OutputPath: "brainstorm.md",
				Context:    map[string]string{},
			},
			[]string{"What is the problem?"},
		)
	})

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs)
	return c, userIn, userOut
}

func TestCoordinator_HITLSingleQuestion(t *testing.T) {
	c, userIn, userOut := setupHITL()

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "I want to build something"

	prompt := <-userOut
	if prompt != "What is the problem?" {
		t.Errorf("got prompt %q, want %q", prompt, "What is the problem?")
	}

	userIn <- "task management"

	msg := <-userOut
	if msg != "Done: brainstorm.md" {
		t.Errorf("got %q, want %q", msg, "Done: brainstorm.md")
	}

	ctx := c.Context()
	if ctx["answer_What is the problem?"] != "task management" {
		t.Errorf("answer = %q, want %q",
			ctx["answer_What is the problem?"], "task management")
	}

	c.Cancel()
	<-done
}

func TestCoordinator_HITLMultipleQuestions(t *testing.T) {
	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewHITLStubAgent(
			agent.Brainstorm,
			&agent.Result{
				OutputPath: "brainstorm.md",
				Context:    map[string]string{},
			},
			[]string{"Q1", "Q2"},
		)
	})

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs)

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"

	prompt := <-userOut
	if prompt != "Q1" {
		t.Errorf("first prompt = %q, want %q", prompt, "Q1")
	}
	userIn <- "A1"

	prompt = <-userOut
	if prompt != "Q2" {
		t.Errorf("second prompt = %q, want %q", prompt, "Q2")
	}
	userIn <- "A2"

	msg := <-userOut
	if msg != "Done: brainstorm.md" {
		t.Errorf("got %q, want done message", msg)
	}

	ctx := c.Context()
	if ctx["answer_Q1"] != "A1" {
		t.Errorf("answer_Q1 = %q, want %q", ctx["answer_Q1"], "A1")
	}
	if ctx["answer_Q2"] != "A2" {
		t.Errorf("answer_Q2 = %q, want %q", ctx["answer_Q2"], "A2")
	}

	c.Cancel()
	<-done
}

// --- Gate ---

type rejectGate struct{}

func (r *rejectGate) Check(
	_ context.Context, _ agent.Role, _ map[string]string,
) (bool, string) {
	return false, "not ready"
}

func TestCoordinator_GateRejects(t *testing.T) {
	regs := agent.NewRegistry()
	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs, WithGate(&rejectGate{}))

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"
	msg := <-userOut
	if msg != "Error: gate rejected brainstorm: not ready" {
		t.Errorf("got %q, want gate rejection message", msg)
	}

	c.Cancel()
	<-done
}

// --- Router ---

type testRouter struct {
	role agent.Role
}

func (tr *testRouter) Route(
	_ context.Context, _ string, _ map[string]string,
) (agent.Role, error) {
	return tr.role, nil
}

func TestCoordinator_RouterDispatches(t *testing.T) {
	regs := agent.NewRegistry()
	regs.Register(agent.Research, func() agent.Agent {
		return agent.NewStubAgent(agent.Research, &agent.Result{
			OutputPath: "research.md",
		})
	})

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs,
		WithRouter(&testRouter{role: agent.Research}),
	)

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"
	msg := <-userOut
	if msg != "Done: research.md" {
		t.Errorf("got %q, want %q", msg, "Done: research.md")
	}

	c.Cancel()
	<-done
}
