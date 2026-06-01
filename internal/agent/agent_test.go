package agent

import (
	"context"
	"testing"

	"github.com/cajohnson0125/Twirl/internal/state"
)

func TestStubAgent_ReturnsResult(t *testing.T) {
	want := &state.Result{
		Status:     state.StatusCompleted,
		OutputPath: "brainstorm.md",
	}
	a := NewStubAgent(Brainstorm, want)

	if a.Role() != Brainstorm {
		t.Errorf("Role: got %q, want %q", a.Role(), Brainstorm)
	}

	got, err := a.Execute(context.Background(), &Task{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got != want {
		t.Errorf("Result: got %+v, want %+v", got, want)
	}
}

func TestStubAgent_ReturnsError(t *testing.T) {
	a := NewStubAgentWithError(Research, context.Canceled)

	_, err := a.Execute(context.Background(), &Task{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("error: got %v, want context.Canceled", err)
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(Brainstorm, func() Agent {
		return NewStubAgent(Brainstorm, &state.Result{})
	})

	a, err := r.Get(Brainstorm)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if a.Role() != Brainstorm {
		t.Errorf("Role: got %q, want %q", a.Role(), Brainstorm)
	}
}

func TestRegistry_GetUnregistered(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get(Execution)
	if err == nil {
		t.Fatal("expected error for unregistered role")
	}
}

func TestRegistry_DuplicateRegisterPanics(t *testing.T) {
	r := NewRegistry()
	r.Register(Brainstorm, func() Agent {
		return NewStubAgent(Brainstorm, &state.Result{})
	})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	r.Register(Brainstorm, func() Agent {
		return NewStubAgent(Brainstorm, &state.Result{})
	})
}

func TestRegistry_Roles(t *testing.T) {
	r := NewRegistry()
	r.Register(Brainstorm, func() Agent {
		return NewStubAgent(Brainstorm, &state.Result{})
	})
	r.Register(Research, func() Agent {
		return NewStubAgent(Research, &state.Result{})
	})

	roles := r.Roles()
	if len(roles) != 2 {
		t.Fatalf("Roles: got %d, want 2", len(roles))
	}

	got := map[Role]bool{}
	for _, r := range roles {
		got[r] = true
	}
	if !got[Brainstorm] || !got[Research] {
		t.Errorf("Roles: got %v, want brainstorm+research", got)
	}
}

func TestTask_Fields(t *testing.T) {
	task := &Task{
		Instruction:  "Brainstorm auth approaches",
		Context:      map[string]string{"topic": "auth"},
		TemplatePath: "(topic)-brainstorm.md",
	}
	if task.Instruction != "Brainstorm auth approaches" {
		t.Errorf("Instruction: got %q", task.Instruction)
	}
	if task.Context["topic"] != "auth" {
		t.Errorf("Context[topic]: got %q", task.Context["topic"])
	}
	if task.TemplatePath != "(topic)-brainstorm.md" {
		t.Errorf("TemplatePath: got %q", task.TemplatePath)
	}
}
