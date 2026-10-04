package mermaid

import (
	"math"
	"slices"
)

const (
	// gapX is the number of columns between boxes side by side, and gapY
	// the number of rows between ranks.
	gapX = 3
	gapY = 2
	// orderSweeps is how many times the ranks are reordered to reduce the
	// crossings of edges, and placeSweeps how many times the boxes are
	// moved toward the boxes they are joined to.
	orderSweeps = 8
	placeSweeps = 10
	// labelGap is the number of columns from the middle of a node to where
	// the label of the edges ending at it starts, leaving a space after the
	// head.
	labelGap = 2
	// backX is the column, left of the middle of a node, that edges going
	// back up leave it or reach it at, so that they keep apart from the
	// edges going down. backWidth is the least width of a box with them,
	// keeping the column off its corners.
	backX     = 2
	backWidth = 2*backX + 3
	// loopHeight is the least height of a box with an edge to itself, which
	// runs out of its right side on one row and back in on the next.
	loopHeight = 4
)

// A layout places the nodes of a flowchart on a grid, in ranks from the
// top down. An edge that spans more than one rank passes through a point,
// a vertex without a node, in each rank between its ends, so that every
// part of an edge joins vertices in adjacent ranks.
type layout struct {
	verts []vertex
	// ranks holds the vertices of each rank from left to right.
	ranks [][]int
	// paths holds the vertices each edge passes through from the top down,
	// which is from its to end to its from end for an edge going back up.
	// Edges from a node to itself have no path.
	paths [][]int
	// back reports for each edge whether it goes back up.
	back []bool
	// parts holds the parts of the edges, each joining vertices in
	// adjacent ranks.
	parts []part
	// rankY and rankH are the top row and the height of each rank. gapTop
	// holds the top row of the gap below each rank but the last, where the
	// parts of edges run sideways in tracks, one row each.
	rankY, rankH, gapTop []int
	w, h                 int
}

type vertex struct {
	// node is the index of the vertex's node, or -1 for a point an edge
	// passes through, edge being the index of that edge.
	node, edge int
	rank       int
	// lines is the wrapped label of a node.
	lines []string
	// x, y, w and h are the box of a node, or the column and the rows a
	// point takes up. cx is the column at the middle of the box.
	x, y, w, h int
	cx         int
	// label is the text written to the right of the head row above a node,
	// labelling the edges that end at it, and backLabel the text written to
	// the left of it, labelling the edges going back up from it.
	label, backLabel string
	// loop reports whether a node has edges to itself, labelled loopLabel.
	loop      bool
	loopLabel string
}

func (v *vertex) isPoint() bool { return v.node < 0 }

// left and right are the number of columns a vertex takes up on each side of
// its middle column, including the labels and the edges to itself beside a
// node.
func (v *vertex) left() int {
	l := v.w / 2
	if v.backLabel != "" {
		l = max(l, backX+1+textWidth(v.backLabel))
	}
	return l
}

func (v *vertex) right() int {
	r := v.w - 1 - v.w/2
	if v.label != "" {
		r = max(r, labelGap-1+textWidth(v.label))
	}
	if v.loop {
		r = max(r, v.loopDX())
		if v.loopLabel != "" {
			r = max(r, v.loopDX()+1+textWidth(v.loopLabel))
		}
	}
	return r
}

// loopDX is how many columns right of its middle the edges from a node to
// itself turn, a column out from its right side.
func (v *vertex) loopDX() int { return v.w - v.w/2 + 1 }

// A part of an edge joins vertices in adjacent ranks. It leaves its from
// vertex at column fromX and reaches its to vertex at column toX, running
// sideways along a track in the gap between the ranks when they differ.
type part struct {
	edge, from, to int
	first, last    bool
	fromX, toX     int
	track          int
}

// layOut places a flowchart's nodes, which must flow from the top down.
func layOut(fc *flowchart) (*layout, error) {
	back := backEdges(fc)
	ranks := rankNodes(fc, back)
	l := &layout{back: back}
	l.addVertices(fc, ranks)
	l.initOrder()
	l.reduceCrossings()
	l.placeX()
	l.addParts()
	l.placeY()
	return l, nil
}

// backEdges finds the edges that go back up, so that the rest have no
// cycles: those that a depth-first walk along the edges, in the order nodes
// and edges are declared, finds leading back to a node it is still walking
// from. Edges from a node to itself are left to be drawn beside it.
func backEdges(fc *flowchart) []bool {
	out := make([][]int, len(fc.nodes))
	for i, e := range fc.edges {
		if e.from != e.to {
			out[e.from] = append(out[e.from], i)
		}
	}
	back := make([]bool, len(fc.edges))
	const (
		unvisited = iota
		walking
		done
	)
	state := make([]int, len(fc.nodes))
	var walk func(u int)
	walk = func(u int) {
		state[u] = walking
		for _, i := range out[u] {
			switch to := fc.edges[i].to; state[to] {
			case unvisited:
				walk(to)
			case walking:
				back[i] = true
			}
		}
		state[u] = done
	}
	for u := range fc.nodes {
		if state[u] == unvisited {
			walk(u)
		}
	}
	return back
}

// A link joins two nodes from the top down, as an edge or an edge going back
// up turned around.
type link struct{ from, to, length int }

// links returns the edges of a flowchart joining two nodes from the top down.
func links(fc *flowchart, back []bool) []link {
	var ls []link
	for i, e := range fc.edges {
		switch {
		case e.from == e.to:
		case back[i]:
			ls = append(ls, link{e.to, e.from, e.length})
		default:
			ls = append(ls, link{e.from, e.to, e.length})
		}
	}
	return ls
}

// rankNodes puts each node in a rank below those of the nodes linked to it,
// by at least the links' lengths, and then moves nodes without links to them
// down next to the nodes they link to.
func rankNodes(fc *flowchart, back []bool) []int {
	n := len(fc.nodes)
	in := make([]int, n)
	out := make([][]link, n)
	for _, lk := range links(fc, back) {
		in[lk.to]++
		out[lk.from] = append(out[lk.from], lk)
	}
	hasIncoming := make([]bool, n)
	for i := range n {
		hasIncoming[i] = in[i] > 0
	}

	// Visit the nodes in topological order, keeping the order they were
	// declared in where it is free.
	var sorted []int
	queue := []int{}
	for i := range n {
		if in[i] == 0 {
			queue = append(queue, i)
		}
	}
	rank := make([]int, n)
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		sorted = append(sorted, u)
		for _, lk := range out[u] {
			rank[lk.to] = max(rank[lk.to], rank[u]+lk.length)
			if in[lk.to]--; in[lk.to] == 0 {
				queue = append(queue, lk.to)
			}
		}
	}

	for _, u := range slices.Backward(sorted) {
		if len(out[u]) == 0 || hasIncoming[u] {
			continue
		}
		lowest := math.MaxInt
		for _, lk := range out[u] {
			lowest = min(lowest, rank[lk.to]-lk.length)
		}
		rank[u] = lowest
	}

	// Moving nodes down may have emptied the top ranks.
	if n > 0 {
		top := slices.Min(rank)
		for i := range rank {
			rank[i] -= top
		}
	}
	return rank
}

// addVertices adds a vertex for each node and the points edges pass
// through.
func (l *layout) addVertices(fc *flowchart, ranks []int) {
	l.ranks = make([][]int, slices.Max(append(ranks, 0))+1)
	add := func(v vertex) int {
		l.verts = append(l.verts, v)
		i := len(l.verts) - 1
		l.ranks[v.rank] = append(l.ranks[v.rank], i)
		return i
	}

	for i, n := range fc.nodes {
		lines := wrapLabel(n.label, labelWidth, labelLines)
		w := 0
		for _, line := range lines {
			w = max(w, textWidth(line))
		}
		add(vertex{node: i, edge: -1, rank: ranks[i], lines: lines, w: w + 2*labelPad + 2, h: len(lines) + 2})
	}
	for ei, e := range fc.edges {
		label := ""
		if e.label != "" && !e.invisible {
			label = fitLabel(e.label, labelWidth)
		}
		if e.from == e.to {
			v := &l.verts[e.from]
			v.loop = v.loop || !e.invisible
			v.loopLabel = joinLabel(v.loopLabel, label)
			v.h = max(v.h, loopHeight)
			l.paths = append(l.paths, nil)
			continue
		}

		top, bottom := e.from, e.to
		if l.back[ei] {
			top, bottom = e.to, e.from
			for _, v := range []int{top, bottom} {
				l.verts[v].w = max(l.verts[v].w, backWidth)
			}
			l.verts[bottom].backLabel = joinLabel(l.verts[bottom].backLabel, label)
		} else {
			l.verts[bottom].label = joinLabel(l.verts[bottom].label, label)
		}
		path := []int{top}
		for r := ranks[top] + 1; r < ranks[bottom]; r++ {
			path = append(path, add(vertex{node: -1, edge: ei, rank: r, w: 1}))
		}
		l.paths = append(l.paths, append(path, bottom))
	}
}

// joinLabel adds label to the labels of edges at the same place.
func joinLabel(labels, label string) string {
	switch {
	case labels == "":
		return label
	case label == "":
		return labels
	}
	return labels + ", " + label
}

// neighbors returns, for each vertex, the vertices it is joined to in the
// rank above and in the rank below. The vertices joined by edges going back
// up come first, so that walking down the edges in order puts them on the
// left, the side those edges leave and reach their nodes on.
func (l *layout) neighbors() (up, down [][]int) {
	up = make([][]int, len(l.verts))
	down = make([][]int, len(l.verts))
	for _, back := range []bool{true, false} {
		for ei, path := range l.paths {
			if l.back[ei] != back {
				continue
			}
			for i := 1; i < len(path); i++ {
				up[path[i]] = append(up[path[i]], path[i-1])
				down[path[i-1]] = append(down[path[i-1]], path[i])
			}
		}
	}
	return up, down
}

// reduceCrossings reorders each rank to reduce the crossings of edges
// between ranks, sorting vertices by the mean position of their neighbors in
// the rank above and then in the rank below, and keeping the best order
// found.
func (l *layout) reduceCrossings() {
	up, down := l.neighbors()
	pos := make([]float64, len(l.verts))
	setPos := func(rank []int) {
		for i, v := range rank {
			pos[v] = float64(i)
		}
	}
	for _, rank := range l.ranks {
		setPos(rank)
	}

	best := cloneRanks(l.ranks)
	bestCrossings := l.crossings()
	for sweep := range orderSweeps {
		ranks, neigh := l.ranks[1:], up
		if sweep%2 == 1 {
			ranks, neigh = slices.Clone(l.ranks[:len(l.ranks)-1]), down
			slices.Reverse(ranks)
		}
		for _, rank := range ranks {
			bary := make(map[int]float64, len(rank))
			for _, v := range rank {
				bary[v] = pos[v]
				if ns := neigh[v]; len(ns) > 0 {
					sum := 0.0
					for _, u := range ns {
						sum += pos[u]
					}
					bary[v] = sum / float64(len(ns))
				}
			}
			slices.SortStableFunc(rank, func(a, b int) int {
				switch {
				case bary[a] < bary[b]:
					return -1
				case bary[a] > bary[b]:
					return 1
				}
				return 0
			})
			setPos(rank)
		}
		if c := l.crossings(); c < bestCrossings {
			best, bestCrossings = cloneRanks(l.ranks), c
		}
	}
	l.ranks = best
}

// initOrder orders each rank by a depth-first walk down the edges from the
// top, so that the vertices an edge leads to start out near each other and
// near the edge's other end.
func (l *layout) initOrder() {
	_, down := l.neighbors()
	visited := make([]bool, len(l.verts))
	for r := range l.ranks {
		l.ranks[r] = l.ranks[r][:0]
	}
	var visit func(v int)
	visit = func(v int) {
		if visited[v] {
			return
		}
		visited[v] = true
		r := l.verts[v].rank
		l.ranks[r] = append(l.ranks[r], v)
		for _, u := range down[v] {
			visit(u)
		}
	}
	// Start from the top ranks, each in the order its vertices were added.
	starts := make([]int, len(l.verts))
	for i := range starts {
		starts[i] = i
	}
	slices.SortStableFunc(starts, func(a, b int) int { return l.verts[a].rank - l.verts[b].rank })
	for _, v := range starts {
		visit(v)
	}
}

func cloneRanks(ranks [][]int) [][]int {
	c := make([][]int, len(ranks))
	for i, r := range ranks {
		c[i] = slices.Clone(r)
	}
	return c
}

// crossings counts the pairs of edge parts between adjacent ranks that cross
// each other.
func (l *layout) crossings() int {
	pos := make([]int, len(l.verts))
	for _, rank := range l.ranks {
		for i, v := range rank {
			pos[v] = i
		}
	}
	type part struct{ top, bottom int }
	parts := make([][]part, len(l.ranks))
	for _, path := range l.paths {
		for i := 1; i < len(path); i++ {
			r := l.verts[path[i-1]].rank
			parts[r] = append(parts[r], part{pos[path[i-1]], pos[path[i]]})
		}
	}
	n := 0
	for _, ps := range parts {
		for i, a := range ps {
			for _, b := range ps[i+1:] {
				if (a.top-b.top)*(a.bottom-b.bottom) < 0 {
					n++
				}
			}
		}
	}
	return n
}

// placeX sets the columns of the vertices: each rank's boxes side by side,
// moved toward the middle of the vertices they are joined to.
func (l *layout) placeX() {
	up, down := l.neighbors()
	center := make([]float64, len(l.verts))
	for _, rank := range l.ranks {
		x := 0.0
		for i, v := range rank {
			if i > 0 {
				x += l.apart(rank[i-1], v)
			}
			center[v] = x
		}
	}
	for sweep := range placeSweeps {
		if sweep%2 == 0 {
			for _, rank := range l.ranks {
				l.relax(rank, up, center)
			}
		} else {
			for _, rank := range slices.Backward(l.ranks) {
				l.relax(rank, down, center)
			}
		}
	}

	left := math.Inf(1)
	for i := range l.verts {
		left = min(left, center[i]-float64(l.verts[i].left()))
	}
	for i := range l.verts {
		v := &l.verts[i]
		v.cx = int(math.Round(center[i] - left))
		v.x = v.cx - v.w/2
		l.w = max(l.w, v.cx+v.right()+1)
	}
}

// apart is the least distance between the middle columns of the vertices a
// and b, side by side in a rank in that order.
func (l *layout) apart(a, b int) float64 {
	return float64(l.verts[a].right() + 1 + gapX + l.verts[b].left())
}

// relax moves the vertices of a rank toward the mean middle column of the
// vertices they are joined to in neigh, keeping them in order and apart.
func (l *layout) relax(rank []int, neigh [][]int, center []float64) {
	n := len(rank)
	want := make([]float64, n)
	for i, v := range rank {
		want[i] = center[v]
		if ns := neigh[v]; len(ns) > 0 {
			sum := 0.0
			for _, u := range ns {
				sum += center[u]
			}
			want[i] = sum / float64(len(ns))
		}
	}

	// Pack the vertices from the left and from the right, each as close as
	// it can be to where it wants to be, and settle between the two.
	fromLeft := slices.Clone(want)
	for i := 1; i < n; i++ {
		fromLeft[i] = max(fromLeft[i], fromLeft[i-1]+l.apart(rank[i-1], rank[i]))
	}
	fromRight := slices.Clone(want)
	for i := n - 2; i >= 0; i-- {
		fromRight[i] = min(fromRight[i], fromRight[i+1]-l.apart(rank[i], rank[i+1]))
	}
	for i, v := range rank {
		center[v] = (fromLeft[i] + fromRight[i]) / 2
	}
	for i := 1; i < n; i++ {
		center[rank[i]] = max(center[rank[i]], center[rank[i-1]]+l.apart(rank[i-1], rank[i]))
	}
}

// addParts splits the edges into parts, and puts the parts that run
// sideways on tracks so that they don't run along each other.
func (l *layout) addParts() {
	for ei, path := range l.paths {
		for i := 1; i < len(path); i++ {
			from, to := &l.verts[path[i-1]], &l.verts[path[i]]
			p := part{
				edge: ei, from: path[i-1], to: path[i],
				first: i == 1, last: i == len(path)-1,
				fromX: from.cx, toX: to.cx, track: -1,
			}
			switch {
			case l.back[ei]:
				if !from.isPoint() {
					p.fromX -= backX
				}
				if !to.isPoint() {
					p.toX -= backX
				}
			// A part that would only jog by a column leaves its from box
			// straight above where it ends instead.
			case !from.isPoint() && abs(p.fromX-p.toX) <= 1 && p.toX > from.x && p.toX < from.x+from.w-1:
				p.fromX = p.toX
			}
			l.parts = append(l.parts, p)
		}
	}

	for r := range len(l.ranks) - 1 {
		var tracks [][]int // the parts on each track
		for i := range l.parts {
			p := &l.parts[i]
			if l.verts[p.from].rank != r || p.fromX == p.toX {
				continue
			}
			t := slices.IndexFunc(tracks, func(track []int) bool {
				return !slices.ContainsFunc(track, func(j int) bool { return clash(p, &l.parts[j]) })
			})
			if t < 0 {
				t = len(tracks)
				tracks = append(tracks, nil)
			}
			tracks[t] = append(tracks[t], i)
			p.track = t
		}
	}
}

// clash reports whether two parts running sideways can't share a track:
// they would run along or right next to each other, and don't leave or reach
// a vertex at the same column.
func clash(a, b *part) bool {
	if a.from == b.from && a.fromX == b.fromX || a.to == b.to && a.toX == b.toX {
		return false
	}
	aLo, aHi := min(a.fromX, a.toX), max(a.fromX, a.toX)
	bLo, bHi := min(b.fromX, b.toX), max(b.fromX, b.toX)
	return aLo <= bHi+1 && bLo <= aHi+1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// placeY sets the rows of the vertices: the ranks one below another, with
// gaps tall enough for the tracks between them, a row where each part runs
// down alone, so that its style shows, and the row of heads above the boxes
// below.
func (l *layout) placeY() {
	tracks := make([]int, len(l.ranks))
	for _, p := range l.parts {
		r := l.verts[p.from].rank
		tracks[r] = max(tracks[r], p.track+1)
	}

	y := 0
	for r, rank := range l.ranks {
		h := 1
		for _, v := range rank {
			h = max(h, l.verts[v].h)
		}
		l.rankY = append(l.rankY, y)
		l.rankH = append(l.rankH, h)
		for _, v := range rank {
			vert := &l.verts[v]
			if vert.isPoint() {
				vert.y, vert.h = y, h
			} else {
				vert.y = y + (h-vert.h)/2
			}
		}
		y += h
		if r < len(l.ranks)-1 {
			l.gapTop = append(l.gapTop, y)
			y += gapY
			if tracks[r] > 0 {
				y += tracks[r]
			}
		}
	}
	l.h = y
}
