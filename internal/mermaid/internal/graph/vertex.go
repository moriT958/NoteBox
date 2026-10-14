package graph

import (
	"notebox/internal/mermaid/internal/diagram"
	"slices"

	"github.com/charmbracelet/x/ansi"
)

// addVertices adds a vertex for each node, the points edges pass through,
// and fillers that keep the places of frames in ranks without other
// vertices in them.
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
	for ei, e := range g.Edges {
		label := ""
		if e.Label != "" && !e.Invisible {
			label = diagram.FitLabel(e.Label, labelWidth)
		}
		if e.From == e.To {
			v := &l.verts[e.From]
			v.loop = v.loop || !e.Invisible
			v.loopLabel = joinLabel(v.loopLabel, label)
			v.h = max(v.h, loopHeight)
			l.paths = append(l.paths, nil)
			continue
		}

		top, bottom := e.From, e.To
		if l.back[ei] {
			top, bottom = e.To, e.From
			for _, v := range []int{top, bottom} {
				l.verts[v].w = max(l.verts[v].w, l.orient.backWidth())
			}
			l.verts[bottom].backLabel = joinLabel(l.verts[bottom].backLabel, label)
		} else {
			l.verts[bottom].label = joinLabel(l.verts[bottom].label, label)
		}
		path := []int{top}
		for r := ranks[top] + 1; r < ranks[bottom]; r++ {
			f := l.pointFrame(nodeFrame[top], nodeFrame[bottom], r)
			path = append(path, add(vertex{node: -1, edge: ei, frame: f, rank: r, w: 1}))
		}
		l.paths = append(l.paths, append(path, bottom))
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
