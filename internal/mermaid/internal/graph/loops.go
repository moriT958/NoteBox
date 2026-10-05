package graph

import (
	"notebox/internal/mermaid/internal/diagram"

	"github.com/charmbracelet/x/ansi"
)

// drawLoops draws the edges from nodes to themselves, out of the right side
// of a node on its first row inside, and back in on the next.
func drawLoops(p *pen, g *Graph) {
	l := p.l
	for i, e := range g.Edges {
		if e.From != e.To || e.Invisible {
			continue
		}
		v := l.verts[e.From]
		right, turn := v.x+v.w-1, v.cx+v.loopDX()
		out, in := v.y+1, v.y+2
		p.across(out, right, turn, e.Style, diagram.RoleEdge)
		p.down(turn, out, in, e.Style, diagram.RoleEdge)
		p.across(in, right+1, turn, e.Style, diagram.RoleEdge)
		if e.ToHead != HeadNone {
			p.head(right+1, in, e.ToHead, dirLeft)
		} else {
			p.across(in, right, right+1, e.Style, diagram.RoleEdge)
		}
		if e.FromHead != HeadNone {
			p.head(right+1, out, e.FromHead, dirLeft)
		}

		if v.loopLabel == "" || i != firstLoop(g, e.From) {
			continue
		}
		if !l.orient.labelsAlong() {
			p.text(turn+2, out, v.loopLabel)
			continue
		}
		// Under a box, the label runs on from past the loop, away from
		// the edges coming in.
		x0, row := p.at(turn, out)
		x1, _ := p.at(turn, in)
		if p.dir == RightLeft {
			p.c.Text(min(x0, x1)-2-ansi.StringWidth(v.loopLabel)+1, row, v.loopLabel, diagram.RoleEdgeLabel)
		} else {
			p.c.Text(max(x0, x1)+2, row, v.loopLabel, diagram.RoleEdgeLabel)
		}
	}
}

// firstLoop returns the index of the first edge from node n to itself, which
// the labels of all of them are drawn with.
func firstLoop(g *Graph, n int) int {
	for i, e := range g.Edges {
		if e.From == n && e.To == n && !e.Invisible {
			return i
		}
	}
	return -1
}
