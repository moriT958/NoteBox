package graph

// backEdges finds the edges that go back up, so that the rest have no
// cycles: those that a depth-first walk along the edges, in the order nodes
// and edges are declared, finds leading back to a node it is still walking
// from. Edges from a node to itself are left to be drawn beside it.
func backEdges(g *Graph) []bool {
	out := make([][]int, len(g.Nodes))
	for i, e := range g.Edges {
		if e.From != e.To {
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

// A link joins two nodes from the top down, as an edge or an edge going back
// up turned around.
type link struct{ from, to, length int }

// links returns the edges of a graph joining two nodes from the top down.
func links(g *Graph, back []bool) []link {
	var ls []link
	for i, e := range g.Edges {
		switch {
		case e.From == e.To:
		case back[i]:
			ls = append(ls, link{e.To, e.From, e.Length})
		default:
			ls = append(ls, link{e.From, e.To, e.Length})
		}
	}
	return ls
}
