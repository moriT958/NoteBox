package graph

import (
	"notebox/internal/mermaid/internal/diagram"
	"slices"

	"github.com/charmbracelet/x/ansi"
)

// addVertices adds a vertex for each node and the points edges pass
// through.
func (l *layout) addVertices(g *Graph, ranks []int) {
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
		add(vertex{node: i, edge: -1, rank: ranks[i], lines: lines, w: w, h: h})
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
