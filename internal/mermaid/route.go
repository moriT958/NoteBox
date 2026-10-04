package mermaid

// The glyphs of edge heads at the ends of edges running down, at the bottom
// end and at the top end.
var (
	downHeads = map[head]string{headArrow: "▼", headCircle: "o", headCross: "×"}
	upHeads   = map[head]string{headArrow: "▲", headCircle: "o", headCross: "×"}
)

// drawEdges draws the edges of a laid out flowchart, and the labels of the
// edges ending at each node. Boxes must be drawn first, so that edges join
// their outlines.
func drawEdges(c *canvas, fc *flowchart, l *layout) {
	// An edge passes straight down through the ranks between its ends.
	for _, v := range l.verts {
		if e := fc.edges[max(v.edge, 0)]; v.isPoint() && !e.invisible {
			c.vline(v.cx, v.y, v.y+v.h-1, e.style, roleEdge)
		}
	}

	for _, p := range l.parts {
		e := fc.edges[p.edge]
		if e.invisible {
			continue
		}
		from, to := l.verts[p.from], l.verts[p.to]
		top := from.y + from.h - 1
		// A head sits in the row above the box it points at, and a part
		// without one joins the box's outline.
		bottom := to.y
		if p.last && e.toHead != headNone {
			bottom = to.y - 1
		}

		if p.fromX == p.toX {
			c.vline(p.fromX, top, bottom, e.style, roleEdge)
		} else {
			track := l.gapTop[from.rank] + p.track
			c.vline(p.fromX, top, track, e.style, roleEdge)
			c.hline(p.fromX, p.toX, track, e.style, roleEdge)
			c.vline(p.toX, track, bottom, e.style, roleEdge)
		}

		if p.last && e.toHead != headNone {
			c.set(p.toX, bottom, downHeads[e.toHead], roleEdge)
		}
		if p.first && e.fromHead != headNone {
			c.set(p.fromX, top, upHeads[e.fromHead], roleEdge)
		}
	}

	for _, v := range l.verts {
		if v.label != "" {
			c.text(v.cx+labelGap, v.y-1, v.label, roleEdgeLabel)
		}
	}
}
