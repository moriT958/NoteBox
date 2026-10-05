package graph

import (
	"math"
	"slices"
)

// rankNodes puts each node in a rank below those of the nodes linked to it,
// by at least the links' lengths, and then moves nodes without links to them
// down next to the nodes they link to.
func rankNodes(g *Graph, back []bool) []int {
	n := len(g.Nodes)
	in := make([]int, n)
	out := make([][]link, n)
	for _, lk := range links(g, back) {
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
