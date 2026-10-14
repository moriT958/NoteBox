package graph

import (
	"notebox/internal/mermaid/internal/diagram"
	"slices"

	"github.com/charmbracelet/x/ansi"
)

// addVertices adds a vertex for each node, the points edges pass through,
// the ports where edges to groups end, and fillers that keep the places of
// frames in ranks without other vertices in them.
func (l *layout) addVertices(g *Graph, ranks, nodeFrame []int) {
	l.ranks = make([][]int, slices.Max(append(ranks, 0))+1)
	add := func(v vertex) int {
		l.verts = append(l.verts, v)
		i := len(l.verts) - 1
		l.ranks[v.rank] = append(l.ranks[v.rank], i)
		return i
	}

	for i, n := range g.Nodes {
		lines := diagram.WrapLabel(n.Label, labelWidth, labelLines)
		w := 0
		for _, line := range lines {
			w = max(w, ansi.StringWidth(line))
		}
		w, h := w+2*labelPad+2, len(lines)+2
		if l.orient.sideways {
			w, h = h, w
		}
		add(vertex{node: i, edge: -1, frame: nodeFrame[i], rank: ranks[i], lines: lines, w: w, h: h})
	}
	ports := map[[2]int]int{} // the ports of each frame, at its top and bottom
	for ei, e := range g.Edges {
		label := ""
		if e.Label != "" && !e.Invisible {
			label = diagram.FitLabel(e.Label, labelWidth)
		}
		top, bottom, ok := l.pathEnds(ei)
		if !ok {
			v := &l.verts[top.node]
			v.loop = v.loop || !e.Invisible
			v.loopLabel = joinLabel(v.loopLabel, label)
			v.h = max(v.h, loopHeight)
			l.paths = append(l.paths, nil)
			continue
		}

		// An end at a frame is the port at its top or its bottom, in the
		// frame's rank nearest the other end, which the edges stop at the
		// outline before.
		vertexOf := func(en end, upper bool) int {
			if !en.isFrame() {
				return en.node
			}
			key := [2]int{en.frame, 0}
			if upper {
				key[1] = 1
			}
			v, ok := ports[key]
			if !ok {
				v = add(vertex{node: -1, edge: ei, frame: en.frame, port: true, rank: l.endRank(en, upper, ranks), w: 1})
				ports[key] = v
			}
			return v
		}
		tv, bv := vertexOf(top, true), vertexOf(bottom, false)
		bottomHead := e.ToHead
		if l.back[ei] {
			bottomHead = e.FromHead
			for _, v := range []int{tv, bv} {
				if vert := &l.verts[v]; !vert.isPoint() || vert.port {
					vert.w = max(vert.w, l.orient.backWidth())
				}
			}
			l.verts[bv].backLabel = joinLabel(l.verts[bv].backLabel, label)
		} else {
			l.verts[bv].label = joinLabel(l.verts[bv].label, label)
		}
		if bottom.isFrame() && !e.Invisible && (bottomHead != HeadNone || label != "") {
			l.frames[bottom.frame].above = true
		}

		path := []int{tv}
		for r := l.verts[tv].rank + 1; r < l.verts[bv].rank; r++ {
			f := l.pointFrame(l.verts[tv].frame, l.verts[bv].frame, r)
			path = append(path, add(vertex{node: -1, edge: ei, frame: f, rank: r, w: 1}))
		}
		l.paths = append(l.paths, append(path, bv))
	}

	filled := make([][]bool, len(l.frames))
	for f, fr := range l.frames {
		filled[f] = make([]bool, fr.bottom-fr.top+1)
	}
	for _, v := range l.verts {
		for f := v.frame; f >= 0; f = l.frames[f].parent {
			filled[f][v.rank-l.frames[f].top] = true
		}
	}
	// Frames nested in a frame fill its ranks too, so fill the innermost
	// first.
	for f, fr := range slices.Backward(l.frames) {
		for r := fr.top; r <= fr.bottom; r++ {
			if filled[f][r-fr.top] {
				continue
			}
			add(vertex{node: -1, edge: -1, frame: f, rank: r, w: 1})
			for p := f; p >= 0; p = l.frames[p].parent {
				filled[p][r-l.frames[p].top] = true
			}
		}
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

// pathEnds returns the ends of an edge from the top down: from its to end
// for an edge going back up. Edges from a node to itself have no path.
func (l *layout) pathEnds(ei int) (top, bottom end, ok bool) {
	top, bottom = l.ends[ei][0], l.ends[ei][1]
	if top == bottom {
		return top, bottom, false
	}
	if l.back[ei] {
		top, bottom = bottom, top
	}
	return top, bottom, true
}

// endRank returns the rank of an end at the top or the bottom of a path: a
// node's rank, or the bottom or the top rank of a frame.
func (l *layout) endRank(e end, upper bool, ranks []int) int {
	switch {
	case !e.isFrame():
		return ranks[e.node]
	case upper:
		return l.frames[e.frame].bottom
	}
	return l.frames[e.frame].top
}
