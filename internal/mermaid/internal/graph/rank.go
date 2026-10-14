package graph

import (
	"math"
	"slices"
)

// An end of an edge in a layout is a node, or a frame that the edge ends at
// the outline of, the other being -1.
type end struct{ node, frame int }

func (e end) isFrame() bool { return e.frame >= 0 }

// A link keeps vertex to of a ranking at least length ranks below vertex
// from. Links between nodes come from edges, and links from a frame's top
// and to its bottom, of no length, keep the frame around its nodes.
type link struct {
	from, to, length int
	edge             bool
}

// A ranking ranks the nodes of a graph, with a vertex for each node and two
// for each frame, above and below the nodes in it, which edges to the frame
// link to. Frames in frames are linked inside their tops and bottoms.
type ranking struct {
	nodes int
	out   [][]link
}

func (rk *ranking) top(f int) int    { return rk.nodes + 2*f }
func (rk *ranking) bottom(f int) int { return rk.nodes + 2*f + 1 }

func (rk *ranking) link(lk link) { rk.out[lk.from] = append(rk.out[lk.from], lk) }

// upper and lower return the vertex of an end at the top and at the bottom
// of a link.
func (rk *ranking) upper(e end) int {
	if e.isFrame() {
		return rk.bottom(e.frame)
	}
	return e.node
}

func (rk *ranking) lower(e end) int {
	if e.isFrame() {
		return rk.top(e.frame)
	}
	return e.node
}

// reaches reports whether vertex to can be reached from vertex from along
// the links.
func (rk *ranking) reaches(from, to int) bool {
	seen := make([]bool, len(rk.out))
	stack := []int{from}
	seen[from] = true
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if u == to {
			return true
		}
		for _, lk := range rk.out[u] {
			if !seen[lk.to] {
				seen[lk.to] = true
				stack = append(stack, lk.to)
			}
		}
	}
	return false
}

// rankNodes puts each node in a rank below those of the nodes linked to it,
// by at least the links' lengths, and then moves nodes without links to them
// down next to the nodes they link to. It sets the ends of the edges and
// which of them go back up.
//
// An edge to a group puts the whole group below the edge's other end, or
// above it when that would make a cycle. When both would, the edge ends at
// the group's first node instead, as Mermaid's edges to groups always do.
func (l *layout) rankNodes(g *Graph, nodeFrame, groupFrame []int) []int {
	n := len(g.Nodes)
	rk := &ranking{nodes: n, out: make([][]link, n+2*len(l.frames))}
	for v, f := range nodeFrame {
		if f >= 0 {
			rk.link(link{from: rk.top(f), to: v})
			rk.link(link{from: v, to: rk.bottom(f)})
		}
	}
	for f, fr := range l.frames {
		if p := fr.parent; p >= 0 {
			rk.link(link{from: rk.top(p), to: rk.top(f)})
			rk.link(link{from: rk.bottom(f), to: rk.bottom(p)})
		}
	}

	// first holds the first node in each frame.
	first := make([]int, len(l.frames))
	for f := range first {
		first[f] = -1
	}
	for v := n - 1; v >= 0; v-- {
		for f := nodeFrame[v]; f >= 0; f = l.frames[f].parent {
			first[f] = v
		}
	}
	endOf := func(i int, group bool) end {
		if group {
			return end{node: -1, frame: groupFrame[i]}
		}
		return end{node: i, frame: -1}
	}

	l.back = backEdges(g)
	l.ends = make([][2]end, len(g.Edges))
	for i, e := range g.Edges {
		from, to := endOf(e.From, e.FromGroup), endOf(e.To, e.ToGroup)
		l.ends[i] = [2]end{from, to}
		if e.FromGroup || e.ToGroup || from == to {
			continue
		}
		if l.back[i] {
			from, to = to, from
		}
		rk.link(link{from: from.node, to: to.node, length: e.Length, edge: true})
	}
	for i, e := range g.Edges {
		if !e.FromGroup && !e.ToGroup {
			continue
		}
		from, to := l.ends[i][0], l.ends[i][1]
		switch {
		case !rk.reaches(rk.lower(to), rk.upper(from)):
		case !rk.reaches(rk.lower(from), rk.upper(to)):
			l.back[i] = true
		default:
			for _, e := range []*end{&from, &to} {
				if e.isFrame() {
					*e = end{node: first[e.frame], frame: -1}
				}
			}
			l.ends[i] = [2]end{from, to}
			if from == to {
				continue
			}
			l.back[i] = rk.reaches(to.node, from.node)
		}
		if l.back[i] {
			from, to = to, from
		}
		rk.link(link{from: rk.upper(from), to: rk.lower(to), length: e.Length, edge: true})
	}
	return rk.rank()
}

// rank ranks the vertices of a ranking, and returns the ranks of the nodes.
func (rk *ranking) rank() []int {
	nv := len(rk.out)
	in := make([]int, nv)
	hasIncoming := make([]bool, nv)
	for _, lks := range rk.out {
		for _, lk := range lks {
			in[lk.to]++
			hasIncoming[lk.to] = hasIncoming[lk.to] || lk.edge
		}
	}

	// Visit the vertices in topological order, keeping the order nodes were
	// declared in where it is free.
	var sorted []int
	queue := []int{}
	for i := range nv {
		if in[i] == 0 {
			queue = append(queue, i)
		}
	}
	rank := make([]int, nv)
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		sorted = append(sorted, u)
		for _, lk := range rk.out[u] {
			rank[lk.to] = max(rank[lk.to], rank[u]+lk.length)
			if in[lk.to]--; in[lk.to] == 0 {
				queue = append(queue, lk.to)
			}
		}
	}

	// Links of frames only keep vertices in place: a vertex without edges to
	// it, but with edges from it, moves down as far as all its links allow.
	for _, u := range slices.Backward(sorted) {
		if hasIncoming[u] || !slices.ContainsFunc(rk.out[u], func(lk link) bool { return lk.edge }) {
			continue
		}
		lowest := math.MaxInt
		for _, lk := range rk.out[u] {
			lowest = min(lowest, rank[lk.to]-lk.length)
		}
		rank[u] = lowest
	}

	// Moving nodes down may have emptied the top ranks.
	rank = rank[:rk.nodes]
	if len(rank) > 0 {
		top := slices.Min(rank)
		for i := range rank {
			rank[i] -= top
		}
	}
	return rank
}
