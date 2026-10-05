package graph

import "notebox/internal/mermaid/diagram"

// drawEdges draws the edges between nodes, leaving their labels and the
// edges from nodes to themselves to drawLabels and drawLoops. Boxes must be
// drawn first, so that edges join their outlines.
func drawEdges(p *pen, g *Graph) {
	l := p.l
	// An edge passes straight down through the ranks between its ends.
	for _, v := range l.verts {
		if v.isPoint() && !g.Edges[v.edge].Invisible {
			p.down(v.cx, v.y, v.y+v.h-1, g.Edges[v.edge].Style, diagram.RoleEdge)
		}
	}

	for _, pt := range l.parts {
		e := g.Edges[pt.edge]
		if e.Invisible {
			continue
		}
		// An edge going back up runs down from its to end.
		topHead, bottomHead := e.FromHead, e.ToHead
		if l.back[pt.edge] {
			topHead, bottomHead = e.ToHead, e.FromHead
		}

		from, to := l.verts[pt.from], l.verts[pt.to]
		top := from.y + from.h - 1
		// A head sits in the row above the box it points at, and a part
		// without one joins the box's outline.
		bottom := to.y
		if pt.last && bottomHead != HeadNone {
			bottom = to.y - 1
		}

		if pt.fromX == pt.toX {
			p.down(pt.fromX, top, bottom, e.Style, diagram.RoleEdge)
		} else {
			track := l.gapTop[from.rank] + pt.track
			p.down(pt.fromX, top, track, e.Style, diagram.RoleEdge)
			p.across(track, pt.fromX, pt.toX, e.Style, diagram.RoleEdge)
			p.down(pt.toX, track, bottom, e.Style, diagram.RoleEdge)
		}

		if pt.last && bottomHead != HeadNone {
			p.head(pt.toX, bottom, bottomHead, dirDown)
		}
		if pt.first && topHead != HeadNone {
			p.head(pt.fromX, top, topHead, dirUp)
		}
	}
}
