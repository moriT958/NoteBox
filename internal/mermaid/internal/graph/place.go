package graph

import (
	"math"
	"slices"

	"github.com/charmbracelet/x/ansi"
)

// left and right are the number of columns a vertex takes up on each side of
// its middle column, including the labels and the edges to itself beside a
// node. When a graph flows sideways, labels run along its edges instead,
// in the gaps between ranks.
func (l *layout) left(v *vertex) int {
	left := v.w / 2
	if v.backLabel != "" && !l.orient.labelsAlong() {
		left = max(left, l.orient.backX()+1+ansi.StringWidth(v.backLabel))
	}
	return left
}

func (l *layout) right(v *vertex) int {
	r := v.w - 1 - v.w/2
	if v.label != "" && !l.orient.labelsAlong() {
		r = max(r, labelGap-1+ansi.StringWidth(v.label))
	}
	if v.loop {
		r = max(r, v.loopDX())
		if v.loopLabel != "" && !l.orient.labelsAlong() {
			r = max(r, v.loopDX()+1+ansi.StringWidth(v.loopLabel))
		}
	}
	return r
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

	var sides [][2]float64
	if len(l.frames) > 0 {
		sides = l.fitFrames(center)
	}

	left := math.Inf(1)
	for i := range l.verts {
		left = min(left, center[i]-float64(l.left(&l.verts[i])))
	}
	for _, s := range sides {
		left = min(left, s[0])
	}
	for i := range l.verts {
		v := &l.verts[i]
		v.cx = int(math.Round(center[i] - left))
		v.x = v.cx - v.w/2
		l.w = max(l.w, v.cx+l.right(v)+1)
	}
	for f, s := range sides {
		fr := &l.frames[f]
		fr.x = int(math.Round(s[0] - left))
		fr.w = int(math.Round(s[1]-left)) - fr.x + 1
		l.w = max(l.w, fr.x+fr.w)
	}
}

// apart is the least distance between the middle columns of the vertices a
// and b, side by side in a rank in that order, leaving room for the sides of
// the frames between them.
func (l *layout) apart(a, b int) float64 {
	fa, fb := l.verts[a].frame, l.verts[b].frame
	lca := l.lca(fa, fb)
	sides := l.depth(fa) + l.depth(fb) - 2*l.depth(lca)
	return float64(l.right(&l.verts[a]) + 1 + l.orient.gapX() + l.left(&l.verts[b]) + sides*l.orient.frameInset())
}

// fitFrames moves the vertices so that each frame is a rectangle around its
// vertices that the vertices outside it keep out of, and returns the left
// and right columns of each frame. The vertices move as little as they can
// from center.
//
// Each vertex, and each side of a frame, must be far enough right of the
// ones left of it, in its rank or, for a frame, in any of its ranks. Those
// constraints form a graph without cycles, since frames are in the same
// order in every rank. Packing everything as far left as the constraints
// allow, and then as far right, gives two layouts that both keep them, and
// so does the one halfway between.
func (l *layout) fitFrames(center []float64) [][2]float64 {
	inset := float64(l.orient.frameInset())
	gap := float64(l.orient.gapX())
	nv := len(l.verts)
	// Variables 0 to nv-1 are the middle columns of the vertices, and
	// nv+2f and nv+2f+1 the left and right sides of frame f.
	leftSide := func(f int) int { return nv + 2*f }
	rightSide := func(f int) int { return nv + 2*f + 1 }
	type constraint struct {
		from, to int
		d        float64
	}
	var cs []constraint
	// atLeast says that variable to is at least d right of variable from.
	atLeast := func(from, to int, d float64) { cs = append(cs, constraint{from, to, d}) }

	for v := range l.verts {
		vert := &l.verts[v]
		if f := vert.frame; f >= 0 {
			atLeast(leftSide(f), v, float64(l.left(vert))+inset)
			atLeast(v, rightSide(f), float64(l.right(vert))+inset)
		}
	}
	for f, fr := range l.frames {
		if p := fr.parent; p >= 0 {
			atLeast(leftSide(p), leftSide(f), inset)
			atLeast(rightSide(f), rightSide(p), inset)
		}
		if !l.orient.sideways {
			atLeast(leftSide(f), rightSide(f), float64(max(fr.titleSpan()-1, 0)))
		}
	}
	// outermost returns the outermost frame that vertex v is in and vertex
	// u isn't, or -1.
	outermost := func(v, u int) int {
		lca := l.lca(l.verts[v].frame, l.verts[u].frame)
		out := -1
		for f := l.verts[v].frame; f != lca; f = l.frames[f].parent {
			out = f
		}
		return out
	}
	for _, rank := range l.ranks {
		for i := 1; i < len(rank); i++ {
			a, b := rank[i-1], rank[i]
			from, to, d := a, b, float64(l.right(&l.verts[a]))+1+gap+float64(l.left(&l.verts[b]))
			if f := outermost(a, b); f >= 0 {
				from, d = rightSide(f), d-float64(l.right(&l.verts[a]))
			}
			if f := outermost(b, a); f >= 0 {
				to, d = leftSide(f), d-float64(l.left(&l.verts[b]))
			}
			atLeast(from, to, d)
		}
	}

	n := nv + 2*len(l.frames)
	out := make([][]int, n)
	in := make([]int, n)
	for i, c := range cs {
		out[c.from] = append(out[c.from], i)
		in[c.to]++
	}
	var order []int
	for v := range n {
		if in[v] == 0 {
			order = append(order, v)
		}
	}
	for i := 0; i < len(order); i++ {
		for _, ci := range out[order[i]] {
			to := cs[ci].to
			if in[to]--; in[to] == 0 {
				order = append(order, to)
			}
		}
	}
	if len(order) != n {
		panic("graph: frames overlap")
	}

	// Vertices start where they are, and frames around them.
	fromLeft := make([]float64, n)
	fromRight := make([]float64, n)
	copy(fromLeft, center)
	copy(fromRight, center)
	unbounded := func(int) (float64, float64) { return math.Inf(-1), math.Inf(1) }
	for f, s := range l.frameSides(slices.Clone(center), unbounded) {
		fromLeft[leftSide(f)], fromRight[leftSide(f)] = s[0], s[0]
		fromLeft[rightSide(f)], fromRight[rightSide(f)] = s[1], s[1]
	}
	for _, v := range order {
		for _, ci := range out[v] {
			c := cs[ci]
			fromLeft[c.to] = max(fromLeft[c.to], fromLeft[v]+c.d)
		}
	}
	for _, v := range slices.Backward(order) {
		for _, ci := range out[v] {
			c := cs[ci]
			fromRight[v] = min(fromRight[v], fromRight[c.to]-c.d)
		}
	}
	for v := range nv {
		center[v] = (fromLeft[v] + fromRight[v]) / 2
	}

	// Each frame's sides then close in on what is in it, as far as the
	// sides found above, which keep the vertices outside out, allow.
	return l.frameSides(center, func(f int) (lo, hi float64) {
		return (fromLeft[leftSide(f)] + fromRight[leftSide(f)]) / 2,
			(fromLeft[rightSide(f)] + fromRight[rightSide(f)]) / 2
	})
}

// frameSides returns the left and right columns of the frames around the
// vertices centered at center: around what is in them, and wider for their
// titles to both sides, as far as the columns from bounds allow, with what
// is in them moved to the middle.
func (l *layout) frameSides(center []float64, bounds func(f int) (lo, hi float64)) [][2]float64 {
	inset := float64(l.orient.frameInset())
	sides := make([][2]float64, len(l.frames))
	inner := make([][2]float64, len(l.frames))
	for f := range inner {
		inner[f] = [2]float64{math.Inf(1), math.Inf(-1)}
	}
	for v := range l.verts {
		vert := &l.verts[v]
		if f := vert.frame; f >= 0 {
			inner[f][0] = min(inner[f][0], center[v]-float64(l.left(vert))-inset)
			inner[f][1] = max(inner[f][1], center[v]+float64(l.right(vert))+inset)
		}
	}
	for f := len(l.frames) - 1; f >= 0; f-- {
		fr := l.frames[f]
		span := 0.0
		if !l.orient.sideways {
			span = float64(max(fr.titleSpan()-1, 0))
		}
		in := inner[f]
		if extra := span - (in[1] - in[0]); extra > 0 {
			lo, hi := bounds(f)
			left := max(lo, min(in[0]-math.Floor(extra/2), hi-span))
			// What is in the frame only meets what is outside it at the
			// frame's sides, so it can move to the middle of the frame.
			l.shiftFrame(f, left+math.Floor(extra/2)-in[0], center, sides)
			in = [2]float64{left, left + span}
		}
		sides[f] = in
		if p := fr.parent; p >= 0 {
			inner[p][0] = min(inner[p][0], sides[f][0]-inset)
			inner[p][1] = max(inner[p][1], sides[f][1]+inset)
		}
	}
	return sides
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

// placeY sets the rows of the vertices and the frames: the ranks one below
// another, with gaps tall enough for the bottoms of the frames that end
// above, the tracks between the ranks, a row where each part runs down
// alone, so that its style shows, the tops of the frames that start below,
// and the row of heads above the boxes below. When a graph flows sideways,
// the gaps also fit the labels that run along the edges into the boxes
// after them, and frames are made wide enough for their titles, which run
// across the rows.
func (l *layout) placeY() {
	tracks := make([]int, len(l.ranks))
	for _, p := range l.parts {
		r := l.verts[p.from].rank
		tracks[r] = max(tracks[r], p.track+1)
	}
	labels := make([]int, len(l.ranks)) // the widest label into each rank
	if l.orient.labelsAlong() {
		for _, v := range l.verts {
			labels[v.rank] = max(labels[v.rank], ansi.StringWidth(v.label), ansi.StringWidth(v.backLabel))
		}
	}
	// pad holds the rows added inside the top and the bottom of each frame
	// to make it wide enough for its title.
	pad := make([][2]int, len(l.frames))
	for {
		l.stackRanks(tracks, labels, pad)
		if !l.orient.sideways {
			return
		}
		// Widen the innermost frame too narrow for its title, which may
		// widen the frames around it too, until all are wide enough.
		narrow := false
		for f := len(l.frames) - 1; f >= 0 && !narrow; f-- {
			if d := l.frames[f].titleSpan() - l.frames[f].h; d > 0 {
				pad[f][0] += d / 2
				pad[f][1] += d - d/2
				narrow = true
			}
		}
		if !narrow {
			return
		}
	}
}

// stackRanks sets the rows of the ranks, their vertices and the frames,
// with the rows of pad inside the frames.
func (l *layout) stackRanks(tracks, labels []int, pad [][2]int) {
	// rise holds the rows from the top of each frame down to the rank it
	// starts at, and fall the rows from the rank it ends at to its bottom,
	// which take in those of the frames nested in it at the same ranks.
	// opens and closes hold the most of them at each rank.
	rise := make([]int, len(l.frames))
	fall := make([]int, len(l.frames))
	opens := make([]int, len(l.ranks))
	closes := make([]int, len(l.ranks))
	for f := len(l.frames) - 1; f >= 0; f-- {
		fr := l.frames[f]
		rise[f], fall[f] = l.orient.frameRise(fr.top), l.orient.frameFall()
		for _, c := range fr.children {
			if l.frames[c].top == fr.top {
				rise[f] = max(rise[f], rise[c])
			}
			if l.frames[c].bottom == fr.bottom {
				fall[f] = max(fall[f], fall[c])
			}
		}
		rise[f] += 1 + pad[f][0]
		fall[f] += 1 + pad[f][1]
		opens[fr.top] = max(opens[fr.top], rise[f])
		closes[fr.bottom] = max(closes[fr.bottom], fall[f])
	}

	l.rankY, l.rankH, l.gapTop = l.rankY[:0], l.rankH[:0], l.gapTop[:0]
	y := opens[0]
	for r, rank := range l.ranks {
		h := 1
		for _, v := range rank {
			h = max(h, l.verts[v].h)
			// The label of a loop under a box runs on past the box's side,
			// within the rank around the box in its middle.
			if vert := l.verts[v]; l.orient.labelsAlong() && vert.loopLabel != "" {
				h = max(h, 2*(loopLabelX+ansi.StringWidth(vert.loopLabel))-vert.h)
			}
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
			y += closes[r]
			l.gapTop = append(l.gapTop, y)
			gap := gapY
			if tracks[r] > 0 {
				gap = tracks[r] + 2
			}
			// A label runs after the tracks and a line, up to the line and
			// head before the box.
			if w := labels[r+1]; w > 0 {
				gap = max(gap, tracks[r]+w+3)
			}
			// The row of heads is in the frames that start below.
			if opens[r+1] > 0 {
				gap--
			}
			y += gap + opens[r+1]
		}
	}
	l.h = y + closes[len(l.ranks)-1]

	for f := range l.frames {
		fr := &l.frames[f]
		fr.y = l.rankY[fr.top] - rise[f]
		fr.h = l.rankY[fr.bottom] + l.rankH[fr.bottom] + fall[f] - fr.y
	}
}

// alongEnd returns the row of the line between the labels along the edges
// into a vertex and the head before it, or before the frames that start
// above it.
func (l *layout) alongEnd(v *vertex) int {
	end := v.y - 2
	for f := v.frame; f >= 0; f = l.frames[f].parent {
		if fr := l.frames[f]; fr.top == v.rank {
			end = min(end, fr.y-1)
		}
	}
	return end
}

// shiftFrame moves what is in frame f by d columns: the vertices and the
// sides of the frames nested in it.
func (l *layout) shiftFrame(f int, d float64, center []float64, sides [][2]float64) {
	if d == 0 {
		return
	}
	for v, vert := range l.verts {
		if l.lca(vert.frame, f) == f {
			center[v] += d
		}
	}
	for c := f + 1; c < len(l.frames); c++ {
		if l.lca(c, f) == f {
			sides[c][0] += d
			sides[c][1] += d
		}
	}
}
