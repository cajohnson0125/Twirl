package workflow

import (
	"strings"
	"testing"

	"github.com/cajohnson0125/Twirl/internal/state"
)

func TestValidate_ValidLinear(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "brainstorm"})
	g.AddNode(&Node{ID: "b", Role: "research"})
	g.AddEdge("a", "b", nil)

	if err := Validate(g); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidate_StartMissing(t *testing.T) {
	g := NewGraph("missing")
	g.AddNode(&Node{ID: "other", Role: "test"})

	err := Validate(g)
	if err == nil {
		t.Fatal("expected error for missing start node")
	}
	if !strings.Contains(err.Error(), "start node") {
		t.Errorf("error: %v", err)
	}
}

func TestValidate_EdgeToMissingNode(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddEdge("a", "nonexistent", nil)

	err := Validate(g)
	if err == nil {
		t.Fatal("expected error for edge to missing node")
	}
	if !strings.Contains(err.Error(), "target not defined") {
		t.Errorf("error: %v", err)
	}
}

func TestValidate_OrphanNode(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddNode(&Node{ID: "orphan", Role: "test"})

	err := Validate(g)
	if err == nil {
		t.Fatal("expected error for orphan node")
	}
	if !strings.Contains(err.Error(), "orphan") {
		t.Errorf("error: %v", err)
	}
}

func TestValidate_NoTerminals(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddNode(&Node{ID: "b", Role: "test"})
	g.AddEdge("a", "b", nil)
	g.AddEdge("b", "a", nil)

	err := Validate(g)
	if err == nil {
		t.Fatal("expected error for no terminals")
	}
	if !strings.Contains(err.Error(), "no terminal") {
		t.Errorf("error: %v", err)
	}
}

func TestValidate_UnreachableTerminal(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddNode(&Node{ID: "b", Role: "test"})
	g.AddNode(&Node{ID: "c", Role: "test"})

	// a -> b (terminal). c is reachable but loops forever.
	g.AddEdge("a", "b", nil)
	g.AddEdge("a", "c", nil)
	g.AddEdge("c", "c", nil)

	err := Validate(g)
	if err == nil {
		t.Fatal("expected error for infinite cycle")
	}
	if !strings.Contains(err.Error(), "no path to any terminal") {
		t.Errorf("error: %v", err)
	}
}

func TestValidate_ConditionalLoop(t *testing.T) {
	// review -> (PASS) -> done
	// review -> (FAIL) -> plan -> review (loop)
	// This is valid because the loop has a conditional exit.
	g := NewGraph("plan")
	g.AddNode(&Node{ID: "plan", Role: "plan"})
	g.AddNode(&Node{ID: "review", Role: "plan_review"})
	g.AddNode(&Node{ID: "done", Role: "scribe"})

	g.AddEdge("plan", "review", nil)
	g.AddEdge("review", "done",
		func(_ *state.State, _ *state.Result) bool {
			return true
		})
	g.AddEdge("review", "plan",
		func(_ *state.State, _ *state.Result) bool {
			return false
		})

	if err := Validate(g); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidate_DefaultGraph(t *testing.T) {
	g := DefaultGraph()
	if err := Validate(g); err != nil {
		t.Fatalf("DefaultGraph validation: %v", err)
	}
}

func TestValidate_EdgesFromUndefinedNode(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddEdge("ghost", "a", nil)

	err := Validate(g)
	if err == nil {
		t.Fatal("expected error for edges from " +
			"undefined node")
	}
	if !strings.Contains(err.Error(), "undefined node") {
		t.Errorf("error: %v", err)
	}
}
