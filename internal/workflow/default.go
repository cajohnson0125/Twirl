package workflow

import (
	"github.com/cajohnson0125/Twirl/internal/agent"
	"github.com/cajohnson0125/Twirl/internal/state"
)

// DefaultGraph constructs the standard Twirl workflow graph.
// This is the 28-step project lifecycle encoded as a directed
// graph of agent dispatch nodes, HITL gates, and conditional
// routing. All pure Go — no config files.
//
// Phases:
//
//	Scoping:   brainstorm -> research -> report -> HITL gate
//	Research:  HITL -> (loop to research) or -> scribe
//	Planning:  plan -> plan_review -> HITL gate
//	Plan loop: HITL -> (loop to plan) or -> scribe
//	Execution: execution (parallel) -> code_review
//	Review:    code_review -> (issues) triage -> assessment
//	           -> HITL -> execution loop
//	           or -> (no issues) scribe
//	Final:     scribe -> done
func DefaultGraph() *Graph {
	g := NewGraph("brainstorm")

	// --- Agent dispatch nodes ---
	g.AddNode(&Node{ID: "brainstorm", Role: string(agent.Brainstorm)})
	g.AddNode(&Node{ID: "research", Role: string(agent.Research)})
	g.AddNode(&Node{ID: "report", Role: string(agent.Report)})

	// HITL gate after scoping: user picks direction.
	g.AddNode(&Node{
		ID:   "scope_gate",
		Role: "hitl_gate",
		Execute: HITLGate(
			"Research complete. What next?",
			[]string{"plan_it", "more_research"},
		),
	})

	// Scribe documents at major checkpoints.
	g.AddNode(&Node{ID: "scribe_scope", Role: string(agent.Scribe)})

	g.AddNode(&Node{ID: "plan", Role: string(agent.Plan)})
	g.AddNode(&Node{ID: "plan_review", Role: string(agent.PlanReview)})

	// HITL gate after planning: user approves or revises.
	g.AddNode(&Node{
		ID:   "plan_gate",
		Role: "hitl_gate",
		Execute: HITLGate(
			"Plan complete. Approve?",
			[]string{"approve", "revise"},
		),
	})

	g.AddNode(&Node{ID: "scribe_plan", Role: string(agent.Scribe)})

	// Execution nodes. The engine's executeNodes runs these
	// in parallel via errgroup when ActiveNodes contains
	// multiple IDs. The planner agent is responsible for
	// setting "execution_count" and "execution_N" context
	// keys to signal how many parallel tasks to fork.
	g.AddNode(&Node{ID: "execution_1", Role: string(agent.Execution)})
	g.AddNode(&Node{ID: "execution_2", Role: string(agent.Execution)})
	g.AddNode(&Node{ID: "execution_3", Role: string(agent.Execution)})

	// Code review after execution completes.
	g.AddNode(&Node{ID: "code_review", Role: string(agent.CodeReview)})

	// Issue handling path.
	g.AddNode(&Node{ID: "triage", Role: string(agent.Triage)})
	g.AddNode(&Node{ID: "assessment", Role: string(agent.Assessment)})

	// HITL gate for fix/defer decisions.
	g.AddNode(&Node{
		ID:   "fix_gate",
		Role: "hitl_gate",
		Execute: HITLGate(
			"Issues found. Fix or defer?",
			[]string{"fix", "defer"},
		),
	})

	// Fix execution (single stream, not parallel).
	g.AddNode(&Node{ID: "execution_fix", Role: string(agent.Execution)})

	// Scribe documents review findings or final docs.
	g.AddNode(&Node{ID: "scribe_review", Role: string(agent.Scribe)})

	// Terminal node: final documentation.
	g.AddNode(&Node{ID: "scribe_final", Role: string(agent.Scribe)})

	// --- Edges ---

	// Scoping phase: linear through brainstorm -> research ->
	// report -> HITL gate.
	g.AddEdge("brainstorm", "research", nil)
	g.AddEdge("research", "report", nil)
	g.AddEdge("report", "scope_gate", nil)

	// Scope gate routing: plan it or loop back to research.
	g.AddEdge("scope_gate", "scribe_scope",
		ChooseRoute("plan_it"))
	g.AddEdge("scope_gate", "research",
		ChooseRoute("more_research"))

	// Research loop -> scribe -> planning.
	g.AddEdge("scribe_scope", "plan", nil)

	// Planning phase: plan -> review -> HITL gate.
	g.AddEdge("plan", "plan_review", nil)
	g.AddEdge("plan_review", "plan_gate",
		ActionIs("PASS"))
	g.AddEdge("plan_review", "plan",
		ActionIs("FAIL"))

	// Plan gate routing: approve or revise.
	g.AddEdge("plan_gate", "scribe_plan",
		ChooseRoute("approve"))
	g.AddEdge("plan_gate", "plan",
		ChooseRoute("revise"))

	// Plan approved -> scribe -> execution fork.
	// The planner agent signals parallelism via context.
	// "execution_count" = "1" (default) routes to
	// execution_1 only.
	// "execution_count" = "2" routes to execution_1 +
	// execution_2.
	// "execution_count" = "3" routes to all three.
	g.AddEdge("scribe_plan", "execution_1", nil)
	g.AddEdge("scribe_plan", "execution_2",
		func(s *state.State, _ *state.Result) bool {
			return s.Context["execution_count"] == "2" ||
				s.Context["execution_count"] == "3"
		})
	g.AddEdge("scribe_plan", "execution_3",
		func(s *state.State, _ *state.Result) bool {
			return s.Context["execution_count"] == "3"
		})

	// Execution -> code review (all paths converge).
	g.AddEdge("execution_1", "code_review", nil)
	g.AddEdge("execution_2", "code_review", nil)
	g.AddEdge("execution_3", "code_review", nil)

	// Code review conditional routing.
	// Issues found -> triage -> assessment -> HITL gate
	// -> fix execution -> code review (loop).
	g.AddEdge("code_review", "triage", HasIssues())
	g.AddEdge("code_review", "scribe_review", NoIssues())

	g.AddEdge("triage", "assessment", nil)
	g.AddEdge("assessment", "fix_gate", nil)

	g.AddEdge("fix_gate", "execution_fix",
		ChooseRoute("fix"))
	g.AddEdge("fix_gate", "scribe_review",
		ChooseRoute("defer"))

	g.AddEdge("execution_fix", "code_review", nil)

	// Final path: review docs -> final docs -> done.
	g.AddEdge("scribe_review", "scribe_final", nil)

	// scribe_final has no outgoing edges = terminal.

	return g
}
