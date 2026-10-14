package graph

import (
	"math"
	"slices"
)

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

	// Gather the vertices of each frame where the walk first put them.
	pos := l.positions()
	l.orderFrames(pos)
	for _, rank := range l.ranks {
		l.arrange(rank, pos)
	}
}

// positions returns the position of each vertex in its rank.
func (l *layout) positions() []float64 {
	pos := make([]float64, len(l.verts))
	for _, rank := range l.ranks {
		for i, v := range rank {
			pos[v] = float64(i)
		}
	}
	return pos
}

// orderFrames sets the keys that order frames side by side to the mean
// position of the vertices in them.
func (l *layout) orderFrames(pos []float64) {
	sum := make([]float64, len(l.frames))
	n := make([]int, len(l.frames))
	for v, vert := range l.verts {
		for f := vert.frame; f >= 0; f = l.frames[f].parent {
			sum[f] += pos[v]
			n[f]++
		}
	}
	for f := range l.frames {
		l.frames[f].key = sum[f] / float64(n[f])
	}
}

// arrange sorts a rank by the keys of its vertices, keeping the vertices of
// each frame together. Frames are sorted among the vertices beside them by
// the mean key of their vertices, and among each other by their own keys.
func (l *layout) arrange(rank []int, key []float64) {
	if len(l.frames) == 0 {
		slices.SortStableFunc(rank, func(a, b int) int { return compareKeys(key[a], key[b]) })
		return
	}
	copy(rank, l.arrangeIn(-1, rank, key))
}

// arrangeIn arranges vertices in frame f, or the whole graph for -1.
func (l *layout) arrangeIn(f int, verts []int, key []float64) []int {
	// An item is a vertex directly in f, or a frame nested directly in f
	// with its vertices.
	type item struct {
		frame int
		verts []int
		key   float64
	}
	var items []item
	at := map[int]int{} // the item of each nested frame
	for _, v := range verts {
		in := l.within(f, l.verts[v].frame)
		if in < 0 {
			// Ports come after what else is in their frames.
			k := key[v]
			if l.verts[v].port {
				k = math.Inf(1)
			}
			items = append(items, item{frame: -1, verts: []int{v}, key: k})
			continue
		}
		i, ok := at[in]
		if !ok {
			i = len(items)
			at[in] = i
			items = append(items, item{frame: in})
		}
		items[i].verts = append(items[i].verts, v)
		items[i].key += key[v]
	}
	for i := range items {
		items[i].key /= float64(len(items[i].verts))
	}
	slices.SortStableFunc(items, func(a, b item) int { return compareKeys(a.key, b.key) })

	// Frames keep their order among each other in the places they sorted
	// into.
	var slots []int
	var frames []item
	for i, it := range items {
		if it.frame >= 0 {
			slots = append(slots, i)
			frames = append(frames, it)
		}
	}
	// Frames with the same key keep the order they were added in, the same
	// in every rank.
	slices.SortFunc(frames, func(a, b item) int {
		if c := compareKeys(l.frames[a.frame].key, l.frames[b.frame].key); c != 0 {
			return c
		}
		return a.frame - b.frame
	})
	for i, slot := range slots {
		items[slot] = frames[i]
	}

	out := make([]int, 0, len(verts))
	for _, it := range items {
		if it.frame < 0 {
			out = append(out, it.verts...)
		} else {
			out = append(out, l.arrangeIn(it.frame, it.verts, key)...)
		}
	}
	return out
}

func compareKeys(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// reduceCrossings reorders each rank to reduce the crossings of edges
// between ranks, sorting vertices by the mean position of their neighbors in
// the rank above and then in the rank below, and keeping the best order
// found. After each sweep, frames are ordered by where their vertices went.
func (l *layout) reduceCrossings() {
	up, down := l.neighbors()
	pos := l.positions()
	bary := make([]float64, len(l.verts))

	best := cloneRanks(l.ranks)
	bestCrossings := l.crossings()
	for sweep := range orderSweeps {
		ranks, neigh := l.ranks[1:], up
		if sweep%2 == 1 {
			ranks, neigh = slices.Clone(l.ranks[:len(l.ranks)-1]), down
			slices.Reverse(ranks)
		}
		for _, rank := range ranks {
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
			l.arrange(rank, bary)
			for i, v := range rank {
				pos[v] = float64(i)
			}
		}
		// The rank a sweep starts from keeps its order, which may have its
		// frames in their old order, so all ranks take the frames' new one.
		l.orderFrames(pos)
		for _, rank := range l.ranks {
			l.arrange(rank, pos)
			for i, v := range rank {
				pos[v] = float64(i)
			}
		}
		if c := l.crossings(); c < bestCrossings {
			best, bestCrossings = cloneRanks(l.ranks), c
		}
	}
	l.ranks = best
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

func cloneRanks(ranks [][]int) [][]int {
	c := make([][]int, len(ranks))
	for i, r := range ranks {
		c[i] = slices.Clone(r)
	}
	return c
}
