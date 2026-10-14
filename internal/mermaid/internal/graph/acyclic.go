package graph

// backEdges finds the edges between nodes that go back up, so that the rest
// have no cycles: those that a depth-first walk along the edges, in the
// order nodes and edges are declared, finds leading back to a node it is
// still walking from. Edges from a node to itself are left to be drawn
// beside it, and edges to groups to rankNodes.
func backEdges(g *Graph) []bool {
	out := make([][]int, len(g.Nodes))
	for i, e := range g.Edges {
		if !e.FromGroup && !e.ToGroup && e.From != e.To {
			out[e.From] = append(out[e.From], i)
		}
	}
	back := make([]bool, len(g.Edges))
	const (
		unvisited = iota
		walking
		done
	)
	state := make([]int, len(g.Nodes))
	var walk func(u int)
	walk = func(u int) {
		state[u] = walking
		for _, i := range out[u] {
			switch to := g.Edges[i].To; state[to] {
			case unvisited:
				walk(to)
			case walking:
				back[i] = true
			}
		}
		state[u] = done
	}
	for u := range g.Nodes {
		if state[u] == unvisited {
			walk(u)
		}
	}
	return back
}
