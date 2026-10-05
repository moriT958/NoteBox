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

	left := math.Inf(1)
	for i := range l.verts {
		left = min(left, center[i]-float64(l.left(&l.verts[i])))
	}
	for i := range l.verts {
		v := &l.verts[i]
		v.cx = int(math.Round(center[i] - left))
		v.x = v.cx - v.w/2
		l.w = max(l.w, v.cx+l.right(v)+1)
	}
}

// apart is the least distance between the middle columns of the vertices a
// and b, side by side in a rank in that order.
func (l *layout) apart(a, b int) float64 {
	return float64(l.right(&l.verts[a]) + 1 + l.orient.gapX() + l.left(&l.verts[b]))
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

// placeY sets the rows of the vertices: the ranks one below another, with
// gaps tall enough for the tracks between them, a row where each part runs
// down alone, so that its style shows, and the row of heads above the boxes
// below. When a graph flows sideways, the gaps also fit the labels that
// run along the edges into the boxes after them.
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

	y := 0
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
			y += gap
		}
	}
	l.h = y
}
