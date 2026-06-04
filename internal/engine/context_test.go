package engine

import (
	"testing"
)

func TestGetContextForState_Coordinator(t *testing.T) {
	e := New()
	ctx := e.GetContextForState()

	if ctx.SystemPrompt != coordinatorSystemPrompt {
		t.Fatal("expected coordinator system prompt")
	}
	if ctx.MemoryScope != "full" {
		t.Fatalf("expected full memory scope, got %s",
			ctx.MemoryScope)
	}
	if len(ctx.Tools) != 1 || ctx.Tools[0] != "spawn_specialist" {
		t.Fatalf("expected spawn_specialist tool, got %v", ctx.Tools)
	}
}

func TestGetContextForState_Specialist(t *testing.T) {
	e := New()
	e.state = StateSpecialistRoom
	ctx := e.GetContextForState()

	if ctx.SystemPrompt != specialistSystemPrompt {
		t.Fatal("expected specialist system prompt")
	}
	if ctx.MemoryScope != "scoped" {
		t.Fatalf("expected scoped memory, got %s",
			ctx.MemoryScope)
	}
	if len(ctx.Tools) != 0 {
		t.Fatalf("expected no tools, got %v", ctx.Tools)
	}
}

func TestGetContextForState_Filing(t *testing.T) {
	e := New()
	e.state = StateFiling
	ctx := e.GetContextForState()

	if ctx.MemoryScope != "none" {
		t.Fatalf("expected no memory scope, got %s",
			ctx.MemoryScope)
	}
}

func TestBuildPrompt_NilBuilder(t *testing.T) {
	e := New()
	result := e.BuildPrompt(nil, "hello")
	if result != "hello" {
		t.Fatalf("expected passthrough, got %q", result)
	}
}
