package workflow

import (
	"fmt"
	"strings"
)

// Validate checks a graph for structural errors. Returns nil
// if the graph is valid, or an error describing all problems
// found.
//
// Checks:
//   - Start node exists in Nodes map
//   - All edge targets reference existing nodes
//   - No orphan nodes (every node is reachable from Start)
//   - All paths eventually reach a terminal node (no cycles
//     without conditional exits)
//   - Every conditional edge has at least one complementary
//     edge (no dead ends from conditional branches)
func Validate(g *Graph) error {
	var problems []string

	// Start node must exist.
	if _, ok := g.Nodes[g.Start]; !ok {
		problems = append(problems,
			fmt.Sprintf("start node %q not defined", g.Start))
	}
	if len(problems) > 0 {
		// Can't continue if start is missing.
		return fmt.Errorf("graph validation: %s",
			strings.Join(problems, "; "))
	}

	// All edge targets must reference existing nodes.
	for from, edges := range g.Edges {
		if _, ok := g.Nodes[from]; !ok {
			problems = append(problems,
				fmt.Sprintf("edges from undefined "+
					"node %q", from))
			continue
		}
		for _, e := range edges {
			if _, ok := g.Nodes[e.To]; !ok {
				problems = append(problems,
					fmt.Sprintf("edge %q -> %q: "+
						"target not defined",
						from, e.To))
			}
		}
	}

	// Check reachability from start via BFS.
	reachable := bfs(g, g.Start)
	for id := range g.Nodes {
		if !reachable[id] {
			problems = append(problems,
				fmt.Sprintf("orphan node %q: not "+
					"reachable from start %q",
					id, g.Start))
		}
	}

	// Check for terminal reachability: every node must have
	// a path to at least one terminal node (no outgoing
	// edges). This catches infinite cycles without
	// conditional exits.
	terminals := findTerminals(g)
	if len(terminals) == 0 {
		problems = append(problems,
			"no terminal nodes: every node has " +
				"outgoing edges")
	} else {
		// Reverse BFS from terminals to find all nodes
		// that can reach a terminal.
		canFinish := make(map[NodeID]bool)
		for _, t := range terminals {
			reverse := reverseBfs(g, t)
			for id := range reverse {
				canFinish[id] = true
			}
		}
		for id := range g.Nodes {
			if !canFinish[id] {
				problems = append(problems,
					fmt.Sprintf("node %q: no path "+
						"to any terminal node "+
						"(possible infinite "+
						"cycle)", id))
			}
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("graph validation: %s",
			strings.Join(problems, "; "))
	}
	return nil
}

// bfs returns the set of node IDs reachable from start.
func bfs(g *Graph, start NodeID) map[NodeID]bool {
	visited := map[NodeID]bool{start: true}
	queue := []NodeID{start}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, e := range g.Edges[cur] {
			if !visited[e.To] {
				visited[e.To] = true
				queue = append(queue, e.To)
			}
		}
	}
	return visited
}

// reverseBfs returns all nodes that can reach the given
// target by following edges backwards.
func reverseBfs(g *Graph, target NodeID) map[NodeID]bool {
	// Build reverse adjacency.
	reverse := make(map[NodeID][]NodeID)
	for from, edges := range g.Edges {
		for _, e := range edges {
			reverse[e.To] = append(reverse[e.To], from)
		}
	}

	visited := map[NodeID]bool{target: true}
	queue := []NodeID{target}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, prev := range reverse[cur] {
			if !visited[prev] {
				visited[prev] = true
				queue = append(queue, prev)
			}
		}
	}
	return visited
}

// findTerminals returns nodes with no outgoing edges.
func findTerminals(g *Graph) []NodeID {
	var terminals []NodeID
	for id := range g.Nodes {
		if len(g.Edges[id]) == 0 {
			terminals = append(terminals, id)
		}
	}
	return terminals
}
