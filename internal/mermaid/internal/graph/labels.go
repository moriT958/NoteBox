package graph

import "github.com/charmbracelet/x/ansi"

// drawLabels writes the labels of the edges ending at each node: beside the
// heads above the node, or, when a graph flows sideways, along the edges
// before their heads, with a line on each side.
func drawLabels(p *pen) {
	l := p.l
	for _, v := range l.verts {
		if v.label != "" {
			if l.orient.labelsAlong() {
				p.textAlong(v.cx, v.y-2-ansi.StringWidth(v.label), v.label)
			} else {
				p.text(v.cx+labelGap, v.y-1, v.label)
			}
		}
		if v.backLabel != "" {
			x := v.cx - l.orient.backX()
			if l.orient.labelsAlong() {
				p.textAlong(x, v.y-2-ansi.StringWidth(v.backLabel), v.backLabel)
			} else {
				p.text(x-1-ansi.StringWidth(v.backLabel), v.y-1, v.backLabel)
			}
		}
	}
}
