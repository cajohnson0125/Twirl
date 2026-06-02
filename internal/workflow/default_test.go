package workflow

import (
	"testing"

	"github.com/cajohnson0125/Twirl/internal/state"
)

func TestDefaultGraph_Structure(t *testing.T) {
	g := DefaultGraph()

	// Must start at brainstorm.
	if g.Start != "brainstorm" {
		t.Errorf("Start: got %q, want %q",
			g.Start, "brainstorm")
	}

	// Must have all expected nodes.
	expected := []NodeID{
		"brainstorm", "research", "report",
		"scope_gate", "scribe_scope",
		"plan", "plan_review", "plan_gate",
		"scribe_plan",
		"execution_1", "execution_2", "execution_3",
		"code_review", "triage", "assessment",
		"fix_gate", "execution_fix",
		"scribe_review", "scribe_final",
	}
	for _, id := range expected {
		if _, ok := g.Nodes[id]; !ok {
			t.Errorf("missing node %q", id)
		}
	}
	if len(g.Nodes) != len(expected) {
		t.Errorf("Nodes: got %d, want %d",
			len(g.Nodes), len(expected))
	}

	// Validate structural integrity.
	if err := Validate(g); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestDefaultGraph_ScopingPath(t *testing.T) {
	g := DefaultGraph()

	// brainstorm -> research -> report -> scope_gate.
	assertPath(t, g, "brainstorm", "research")
	assertPath(t, g, "research", "report")
	assertPath(t, g, "report", "scope_gate")

	// scope_gate has two conditional edges.
	edges := g.Edges["scope_gate"]
	if len(edges) != 2 {
		t.Fatalf("scope_gate edges: got %d, want 2",
			len(edges))
	}

	// "plan_it" -> scribe_scope.
	found := false
	for _, e := range edges {
		if e.To == "scribe_scope" {
			found = true
			if e.Condition == nil {
				t.Error("scribe_scope edge: " +
					"expected condition")
			}
		}
	}
	if !found {
		t.Error("scope_gate -> scribe_scope not found")
	}

	// "more_research" -> research (loop-back).
	found = false
	for _, e := range edges {
		if e.To == "research" {
			found = true
		}
	}
	if !found {
		t.Error("scope_gate -> research loop-back " +
			"not found")
	}
}

func TestDefaultGraph_PlanningLoop(t *testing.T) {
	g := DefaultGraph()

	// plan -> plan_review.
	assertPath(t, g, "plan", "plan_review")

	// plan_review has two conditional edges:
	// PASS -> plan_gate, FAIL -> plan (loop).
	edges := g.Edges["plan_review"]
	if len(edges) != 2 {
		t.Fatalf("plan_review edges: got %d, want 2",
			len(edges))
	}

	foundPass := false
	foundFail := false
	for _, e := range edges {
		if e.To == "plan_gate" {
			foundPass = true
		}
		if e.To == "plan" {
			foundFail = true
		}
	}
	if !foundPass {
		t.Error("plan_review -> plan_gate not found")
	}
	if !foundFail {
		t.Error("plan_review -> plan loop-back not found")
	}

	// plan_gate has two conditionals: approve/revise.
	edges = g.Edges["plan_gate"]
	if len(edges) != 2 {
		t.Fatalf("plan_gate edges: got %d, want 2",
			len(edges))
	}
}

func TestDefaultGraph_ExecutionParallel(t *testing.T) {
	g := DefaultGraph()

	// scribe_plan -> execution_1 (first exec node).
	assertPath(t, g, "scribe_plan", "execution_1")

	// All execution nodes route to code_review.
	for _, id := range []NodeID{
		"execution_1", "execution_2", "execution_3",
	} {
		found := false
		for _, e := range g.Edges[id] {
			if e.To == "code_review" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s -> code_review not found", id)
		}
	}
}

func TestDefaultGraph_ReviewCycle(t *testing.T) {
	g := DefaultGraph()

	// code_review has two conditional edges:
	// issues -> triage, no issues -> scribe_review.
	edges := g.Edges["code_review"]
	if len(edges) != 2 {
		t.Fatalf("code_review edges: got %d, want 2",
			len(edges))
	}

	foundTriage := false
	foundScribe := false
	for _, e := range edges {
		if e.To == "triage" {
			foundTriage = true
		}
		if e.To == "scribe_review" {
			foundScribe = true
		}
	}
	if !foundTriage {
		t.Error("code_review -> triage not found")
	}
	if !foundScribe {
		t.Error("code_review -> scribe_review not found")
	}

	// Fix loop: triage -> assessment -> fix_gate.
	assertPath(t, g, "triage", "assessment")
	assertPath(t, g, "assessment", "fix_gate")

	// fix_gate: fix -> execution_fix -> code_review (loop).
	// fix_gate: defer -> scribe_review.
	edges = g.Edges["fix_gate"]
	if len(edges) != 2 {
		t.Fatalf("fix_gate edges: got %d, want 2",
			len(edges))
	}
	assertPath(t, g, "execution_fix", "code_review")
}

func TestDefaultGraph_TerminalNode(t *testing.T) {
	g := DefaultGraph()

	// scribe_final has no outgoing edges.
	if len(g.Edges["scribe_final"]) != 0 {
		t.Errorf("scribe_final: got %d edges, want 0",
			len(g.Edges["scribe_final"]))
	}

	// scribe_review -> scribe_final -> terminal.
	assertPath(t, g, "scribe_review", "scribe_final")
}

func TestDefaultGraph_HITLNodes(t *testing.T) {
	g := DefaultGraph()

	// All HITL gate nodes must have Execute set.
	hitlNodes := []NodeID{
		"scope_gate", "plan_gate", "fix_gate",
	}
	for _, id := range hitlNodes {
		node, ok := g.Nodes[id]
		if !ok {
			t.Errorf("node %q not found", id)
			continue
		}
		if node.Execute == nil {
			t.Errorf("node %q: Execute is nil", id)
		}
		if node.Role != "hitl_gate" {
			t.Errorf("node %q: Role = %q, want %q",
				id, node.Role, "hitl_gate")
		}
	}
}

func TestDefaultGraph_HITLGatesReturnPaused(t *testing.T) {
	g := DefaultGraph()

	hitlNodes := []NodeID{
		"scope_gate", "plan_gate", "fix_gate",
	}
	for _, id := range hitlNodes {
		node := g.Nodes[id]
		res, err := node.Execute(nil, &state.State{})
		if err != nil {
			t.Errorf("node %q Execute: %v", id, err)
			continue
		}
		if res.Status != state.StatusPausedHITL {
			t.Errorf("node %q: Status = %d, want %d",
				id, res.Status,
				state.StatusPausedHITL)
		}
		if res.HITLRequest == nil {
			t.Errorf("node %q: HITLRequest is nil", id)
		}
	}
}

func assertPath(
	t *testing.T,
	g *Graph,
	from, to NodeID,
) {
	t.Helper()
	for _, e := range g.Edges[from] {
		if e.To == to {
			return
		}
	}
	t.Errorf("no edge from %q to %q", from, to)
}
