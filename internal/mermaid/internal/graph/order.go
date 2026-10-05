package graph

import "slices"

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
