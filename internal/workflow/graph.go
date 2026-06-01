package workflow

import (
	"context"

	"github.com/cajohnson0125/Twirl/internal/state"
)

// NodeID identifies a node in the workflow graph.
type NodeID string

// NodeFunc is the contract every graph node must satisfy. It
// receives the current project state and returns a result the
// engine uses for routing decisions.
type NodeFunc func(
	ctx context.Context,
	s *state.State,
) (*state.Result, error)

// EdgeCondition evaluates the current state and the result of the
// node that just executed. Returns true if this edge should be
// traversed. A nil condition means unconditional.
type EdgeCondition func(
	s *state.State,
	r *state.Result,
) bool

// Node represents a single step in the workflow — an agent
// dispatch, a HITL gate, or a join point for parallel branches.
type Node struct {
	ID      NodeID
	Role    string // agent role name, "hitl_gate", "join", etc.
	Execute NodeFunc
}

// Edge represents a directed transition between two nodes.
type Edge struct {
	To        NodeID
	Condition EdgeCondition
}

// Graph is a directed graph of nodes and edges. The engine starts
// at Start and walks edges based on conditions. Constructed
// entirely in Go at compile time — no external config files.
type Graph struct {
	Start NodeID
	Nodes map[NodeID]*Node
	Edges map[NodeID][]Edge
}

// NewGraph creates an empty graph with the given start node.
func NewGraph(start NodeID) *Graph {
	return &Graph{
		Start: start,
		Nodes: make(map[NodeID]*Node),
		Edges: make(map[NodeID][]Edge),
	}
}

// AddNode adds a node to the graph. Panics on duplicate ID.
func (g *Graph) AddNode(n *Node) {
	if _, ok := g.Nodes[n.ID]; ok {
		panic("workflow: duplicate node ID: " + n.ID)
	}
	g.Nodes[n.ID] = n
}

// AddEdge adds a directed edge. If cond is nil the edge is
// unconditional.
func (g *Graph) AddEdge(from, to NodeID, cond EdgeCondition) {
	g.Edges[from] = append(g.Edges[from], Edge{
		To:        to,
		Condition: cond,
	})
}
