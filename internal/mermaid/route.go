package mermaid

// headGlyphs are the glyphs of edge heads by the direction they point in.
var headGlyphs = map[dirs]map[head]string{
	dirDown:  {headArrow: "▼", headCircle: "o", headCross: "×"},
	dirUp:    {headArrow: "▲", headCircle: "o", headCross: "×"},
	dirLeft:  {headArrow: "◀", headCircle: "o", headCross: "×"},
	dirRight: {headArrow: "▶", headCircle: "o", headCross: "×"},
}

// A pen draws a layout onto a canvas, turning the layout's positions, which
// are as they would be in a flowchart that runs down, to the direction of
// the flowchart.
type pen struct {
	c   *canvas
	dir flowDirection
	l   *layout
}

func newPen(dir flowDirection, l *layout) *pen {
	w, h := l.w, l.h
	if l.sideways {
		w, h = h, w
	}
	return &pen{c: newCanvas(w, h), dir: dir, l: l}
}

// at returns the cell of the canvas at the position (x, y) of the layout.
func (p *pen) at(x, y int) (int, int) {
	switch p.dir {
	case flowBottomUp:
		return x, p.l.h - 1 - y
	case flowLeftRight:
		return y, x
	case flowRightLeft:
		return p.l.h - 1 - y, x
	}
	return x, y
}

// turn returns the direction on the canvas of a direction in the layout.
func (p *pen) turn(d dirs) dirs {
	turned := map[flowDirection]map[dirs]dirs{
		flowTopDown:   {dirUp: dirUp, dirDown: dirDown, dirLeft: dirLeft, dirRight: dirRight},
		flowBottomUp:  {dirUp: dirDown, dirDown: dirUp, dirLeft: dirLeft, dirRight: dirRight},
		flowLeftRight: {dirUp: dirLeft, dirDown: dirRight, dirLeft: dirUp, dirRight: dirDown},
		flowRightLeft: {dirUp: dirRight, dirDown: dirLeft, dirLeft: dirUp, dirRight: dirDown},
	}[p.dir]
	var out dirs
	for _, one := range []dirs{dirUp, dirDown, dirLeft, dirRight} {
		if d&one != 0 {
			out |= turned[one]
		}
	}
	return out
}

// down draws a line in column x of the layout from row y0 to row y1.
func (p *pen) down(x, y0, y1 int, style lineStyle, r role) {
	x0, cy0 := p.at(x, y0)
	x1, cy1 := p.at(x, y1)
	if p.l.sideways {
		p.c.hline(x0, x1, cy0, style, r)
	} else {
		p.c.vline(x0, cy0, cy1, style, r)
	}
}

// across draws a line in row y of the layout from column x0 to column x1.
func (p *pen) across(y, x0, x1 int, style lineStyle, r role) {
	cx0, cy0 := p.at(x0, y)
	cx1, cy1 := p.at(x1, y)
	if p.l.sideways {
		p.c.vline(cx0, cy0, cy1, style, r)
	} else {
		p.c.hline(cx0, cx1, cy0, style, r)
	}
}

// join adds a line leaving (x, y) of the layout in the directions d.
func (p *pen) join(x, y int, d dirs, style lineStyle) {
	cx, cy := p.at(x, y)
	p.c.line(cx, cy, p.turn(d), style, roleEdge)
}

// head draws a head at (x, y) of the layout, pointing in the direction d.
func (p *pen) head(x, y int, h head, d dirs) {
	cx, cy := p.at(x, y)
	p.c.set(cx, cy, headGlyphs[p.turn(d)][h], roleEdge)
}

// box draws the box of a node with its label.
func (p *pen) box(v vertex, sh shape) {
	x0, y0 := p.at(v.x, v.y)
	x1, y1 := p.at(v.x+v.w-1, v.y+v.h-1)
	x0, x1 = min(x0, x1), max(x0, x1)
	y0, y1 = min(y0, y1), max(y0, y1)
	drawBox(p.c, x0, y0, x1-x0+1, y1-y0+1, sh, v.lines)
}

// text writes s from (x, y) of the layout to the right, for flowcharts that
// run up or down.
func (p *pen) text(x, y int, s string) {
	cx, cy := p.at(x, y)
	p.c.text(cx, cy, s, roleEdgeLabel)
}

// textAlong writes s along column x of the layout, in the rows from y to y
// plus its width, for flowcharts that run sideways, whose edges run along
// the canvas's rows there.
func (p *pen) textAlong(x, y int, s string) {
	cx0, cy := p.at(x, y)
	cx1, _ := p.at(x, y+textWidth(s)-1)
	p.c.text(min(cx0, cx1), cy, s, roleEdgeLabel)
}

// drawEdges draws the edges of a laid out flowchart, and the labels of the
// edges at each node. Boxes must be drawn first, so that edges join their
// outlines.
func drawEdges(p *pen, fc *flowchart) {
	l := p.l
	// An edge passes straight down through the ranks between its ends.
	for _, v := range l.verts {
		if v.isPoint() && !fc.edges[v.edge].invisible {
			p.down(v.cx, v.y, v.y+v.h-1, fc.edges[v.edge].style, roleEdge)
		}
	}

	for _, pt := range l.parts {
		e := fc.edges[pt.edge]
		if e.invisible {
			continue
		}
		// An edge going back up runs down from its to end.
		topHead, bottomHead := e.fromHead, e.toHead
		if l.back[pt.edge] {
			topHead, bottomHead = e.toHead, e.fromHead
		}

		from, to := l.verts[pt.from], l.verts[pt.to]
		top := from.y + from.h - 1
		// A head sits in the row above the box it points at, and a part
		// without one joins the box's outline.
		bottom := to.y
		if pt.last && bottomHead != headNone {
			bottom = to.y - 1
		}

		if pt.fromX == pt.toX {
			p.down(pt.fromX, top, bottom, e.style, roleEdge)
		} else {
			track := l.gapTop[from.rank] + pt.track
			p.down(pt.fromX, top, track, e.style, roleEdge)
			p.across(track, pt.fromX, pt.toX, e.style, roleEdge)
			p.down(pt.toX, track, bottom, e.style, roleEdge)
		}

		if pt.last && bottomHead != headNone {
			p.head(pt.toX, bottom, bottomHead, dirDown)
		}
		if pt.first && topHead != headNone {
			p.head(pt.fromX, top, topHead, dirUp)
		}
	}

	drawLabels(p)
	drawLoops(p, fc)
}

// drawLabels writes the labels of the edges ending at each node: beside the
// heads above the node, or, when a flowchart runs sideways, along the edges
// before their heads, with a line on each side.
func drawLabels(p *pen) {
	l := p.l
	for _, v := range l.verts {
		if v.label != "" {
			if l.sideways {
				p.textAlong(v.cx, v.y-2-textWidth(v.label), v.label)
			} else {
				p.text(v.cx+labelGap, v.y-1, v.label)
			}
		}
		if v.backLabel != "" {
			x := v.cx - l.backX()
			if l.sideways {
				p.textAlong(x, v.y-2-textWidth(v.backLabel), v.backLabel)
			} else {
				p.text(x-1-textWidth(v.backLabel), v.y-1, v.backLabel)
			}
		}
	}
}

// drawLoops draws the edges from nodes to themselves, out of the right side
// of a node on its first row inside, and back in on the next.
func drawLoops(p *pen, fc *flowchart) {
	l := p.l
	for i, e := range fc.edges {
		if e.from != e.to || e.invisible {
			continue
		}
		v := l.verts[e.from]
		right, turn := v.x+v.w-1, v.cx+v.loopDX()
		out, in := v.y+1, v.y+2
		p.across(out, right, turn, e.style, roleEdge)
		p.down(turn, out, in, e.style, roleEdge)
		p.across(in, right+1, turn, e.style, roleEdge)
		if e.toHead != headNone {
			p.head(right+1, in, e.toHead, dirLeft)
		} else {
			p.join(right, in, dirRight, e.style)
		}
		if e.fromHead != headNone {
			p.head(right+1, out, e.fromHead, dirLeft)
		}

		if v.loopLabel == "" || i != firstLoop(fc, e.from) {
			continue
		}
		if !l.sideways {
			p.text(turn+2, out, v.loopLabel)
			continue
		}
		// Under a box, the label runs on from past the loop, away from
		// the edges coming in.
		x0, row := p.at(turn, out)
		x1, _ := p.at(turn, in)
		if p.dir == flowRightLeft {
			p.c.text(min(x0, x1)-2-textWidth(v.loopLabel)+1, row, v.loopLabel, roleEdgeLabel)
		} else {
			p.c.text(max(x0, x1)+2, row, v.loopLabel, roleEdgeLabel)
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
