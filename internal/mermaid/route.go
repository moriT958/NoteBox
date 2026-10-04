package mermaid

// The glyphs of edge heads at the ends of edges running down, at the bottom
// end and at the top end, and of edges to a node itself, which run back
// into the node's right side.
var (
	downHeads = map[head]string{headArrow: "▼", headCircle: "o", headCross: "×"}
	upHeads   = map[head]string{headArrow: "▲", headCircle: "o", headCross: "×"}
	leftHeads = map[head]string{headArrow: "◀", headCircle: "o", headCross: "×"}
)

// drawEdges draws the edges of a laid out flowchart, and the labels of the
// edges at each node. Boxes must be drawn first, so that edges join their
// outlines.
func drawEdges(c *canvas, fc *flowchart, l *layout) {
	// An edge passes straight down through the ranks between its ends.
	for _, v := range l.verts {
		if v.isPoint() && !fc.edges[v.edge].invisible {
			c.vline(v.cx, v.y, v.y+v.h-1, fc.edges[v.edge].style, roleEdge)
		}
	}

	for _, p := range l.parts {
		e := fc.edges[p.edge]
		if e.invisible {
			continue
		}
		// An edge going back up runs down from its to end.
		topHead, bottomHead := e.fromHead, e.toHead
		if l.back[p.edge] {
			topHead, bottomHead = e.toHead, e.fromHead
		}

		from, to := l.verts[p.from], l.verts[p.to]
		top := from.y + from.h - 1
		// A head sits in the row above the box it points at, and a part
		// without one joins the box's outline.
		bottom := to.y
		if p.last && bottomHead != headNone {
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

		if p.last && bottomHead != headNone {
			c.set(p.toX, bottom, downHeads[bottomHead], roleEdge)
		}
		if p.first && topHead != headNone {
			c.set(p.fromX, top, upHeads[topHead], roleEdge)
		}
	}

	for _, v := range l.verts {
		if v.label != "" {
			c.text(v.cx+labelGap, v.y-1, v.label, roleEdgeLabel)
		}
		if v.backLabel != "" {
			c.text(v.cx-backX-1-textWidth(v.backLabel), v.y-1, v.backLabel, roleEdgeLabel)
		}
	}
	drawLoops(c, fc, l)
}

// drawLoops draws the edges from nodes to themselves, out of the right side
// of a node on its first row inside, and back in on the next.
func drawLoops(c *canvas, fc *flowchart, l *layout) {
	for i, e := range fc.edges {
		if e.from != e.to || e.invisible {
			continue
		}
		v := l.verts[e.from]
		right, turn := v.x+v.w-1, v.cx+v.loopDX()
		out, in := v.y+1, v.y+2
		c.hline(right, turn, out, e.style, roleEdge)
		c.vline(turn, out, in, e.style, roleEdge)
		c.hline(right+1, turn, in, e.style, roleEdge)
		if e.toHead != headNone {
			c.set(right+1, in, leftHeads[e.toHead], roleEdge)
		} else {
			c.line(right, in, dirRight, e.style, roleEdge)
		}
		if e.fromHead != headNone {
			c.set(right+1, out, leftHeads[e.fromHead], roleEdge)
		}
		if v.loopLabel != "" && i == firstLoop(fc, e.from) {
			c.text(turn+2, out, v.loopLabel, roleEdgeLabel)
		}
	}
}

// firstLoop returns the index of the first edge from node n to itself, which
// the labels of all of them are drawn with.
func firstLoop(fc *flowchart, n int) int {
	for i, e := range fc.edges {
		if e.from == n && e.to == n && !e.invisible {
			return i
		}
	}
	return -1
}
