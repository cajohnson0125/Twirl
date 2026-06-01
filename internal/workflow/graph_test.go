package workflow

import (
	"context"
	"testing"

	"github.com/cajohnson0125/Twirl/internal/state"
)

func TestNewGraph_SetsStart(t *testing.T) {
	g := NewGraph("start")
	if g.Start != "start" {
		t.Errorf("Start: got %q, want %q", g.Start, "start")
	}
	if len(g.Nodes) != 0 {
		t.Errorf("Nodes: expected empty, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 0 {
		t.Errorf("Edges: expected empty, got %d", len(g.Edges))
	}
}

func TestAddNode(t *testing.T) {
	g := NewGraph("a")
	n := &Node{
		ID:   "a",
		Role: "brainstorm",
		Execute: func(
			ctx context.Context,
			s *state.State,
		) (*state.Result, error) {
			return &state.Result{}, nil
		},
	}
	g.AddNode(n)

	if len(g.Nodes) != 1 {
		t.Fatalf("Nodes: got %d, want 1", len(g.Nodes))
	}
	if g.Nodes["a"] != n {
		t.Error("node not stored correctly")
	}
}

func TestAddNode_DuplicatePanics(t *testing.T) {
	g := NewGraph("a")
	n := &Node{ID: "a", Role: "test"}
	g.AddNode(n)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate node ID")
		}
	}()
	g.AddNode(&Node{ID: "a", Role: "other"})
}

func TestAddEdge(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddNode(&Node{ID: "b", Role: "test"})

	g.AddEdge("a", "b", nil)

	if len(g.Edges["a"]) != 1 {
		t.Fatalf("Edges[a]: got %d, want 1", len(g.Edges["a"]))
	}
	edge := g.Edges["a"][0]
	if edge.To != "b" {
		t.Errorf("edge.To: got %q, want %q", edge.To, "b")
	}
	if edge.Condition != nil {
		t.Error("edge.Condition: expected nil for unconditional")
	}
}

func TestAddEdge_WithCondition(t *testing.T) {
	g := NewGraph("a")
	g.AddNode(&Node{ID: "a", Role: "test"})
	g.AddNode(&Node{ID: "b", Role: "test"})
	g.AddNode(&Node{ID: "c", Role: "test"})

	cond := func(s *state.State, r *state.Result) bool {
		return r.Severity > 0
	}
	g.AddEdge("a", "b", cond)
	g.AddEdge("a", "c", nil)

	if len(g.Edges["a"]) != 2 {
		t.Fatalf("Edges[a]: got %d, want 2", len(g.Edges["a"]))
	}
	if g.Edges["a"][0].Condition == nil {
		t.Error("first edge condition: expected non-nil")
	}
	if g.Edges["a"][1].Condition != nil {
		t.Error("second edge condition: expected nil")
	}
}

func TestAddEdge_MultipleFromSameNode(t *testing.T) {
	g := NewGraph("start")
	for _, id := range []string{"start", "a", "b", "c"} {
		g.AddNode(&Node{ID: NodeID(id), Role: "test"})
	}

	g.AddEdge("start", "a", nil)
	g.AddEdge("start", "b", nil)
	g.AddEdge("start", "c", nil)

	if len(g.Edges["start"]) != 3 {
		t.Errorf("Edges[start]: got %d, want 3",
			len(g.Edges["start"]))
	}
}
