package graph

import (
	"notebox/internal/mermaid/internal/diagram"

	"github.com/charmbracelet/x/ansi"
)

// dir is a set of directions on a grid, of the layout or of the canvas.
type dir uint8

const (
	dirUp dir = 1 << iota
	dirDown
	dirLeft
	dirRight
)

// headGlyphs are the glyphs of edge heads by the direction they point in.
var headGlyphs = map[dir]map[Head]string{
	dirDown:  {HeadArrow: "▼", HeadCircle: "o", HeadCross: "×"},
	dirUp:    {HeadArrow: "▲", HeadCircle: "o", HeadCross: "×"},
	dirLeft:  {HeadArrow: "◀", HeadCircle: "o", HeadCross: "×"},
	dirRight: {HeadArrow: "▶", HeadCircle: "o", HeadCross: "×"},
}

// A pen draws a layout onto a canvas, turning the layout's positions, which
// are as they would be in a graph that flows down, to the direction of
// the graph.
type pen struct {
	c   *diagram.Canvas
	dir Direction
	l   *layout
}

func newPen(dir Direction, l *layout) *pen {
	w, h := l.w, l.h
	if l.orient.sideways {
		w, h = h, w
	}
	return &pen{c: diagram.New(w, h), dir: dir, l: l}
}

// at returns the cell of the canvas at the position (x, y) of the layout.
func (p *pen) at(x, y int) (int, int) {
	switch p.dir {
	case BottomUp:
		return x, p.l.h - 1 - y
	case LeftRight:
		return y, x
	case RightLeft:
		return p.l.h - 1 - y, x
	}
	return x, y
}

// turn returns the direction on the canvas of a direction in the layout.
func (p *pen) turn(d dir) dir {
	turned := map[Direction]map[dir]dir{
		TopDown:   {dirUp: dirUp, dirDown: dirDown, dirLeft: dirLeft, dirRight: dirRight},
		BottomUp:  {dirUp: dirDown, dirDown: dirUp, dirLeft: dirLeft, dirRight: dirRight},
		LeftRight: {dirUp: dirLeft, dirDown: dirRight, dirLeft: dirUp, dirRight: dirDown},
		RightLeft: {dirUp: dirRight, dirDown: dirLeft, dirLeft: dirUp, dirRight: dirDown},
	}[p.dir]
	var out dir
	for _, one := range []dir{dirUp, dirDown, dirLeft, dirRight} {
		if d&one != 0 {
			out |= turned[one]
		}
	}
	return out
}

// down draws a line in column x of the layout from row y0 to row y1.
func (p *pen) down(x, y0, y1 int, style diagram.LineStyle, r diagram.Role) {
	x0, cy0 := p.at(x, y0)
	x1, cy1 := p.at(x, y1)
	if p.l.orient.sideways {
		p.c.HLine(x0, x1, cy0, style, r)
	} else {
		p.c.VLine(x0, cy0, cy1, style, r)
	}
}

// across draws a line in row y of the layout from column x0 to column x1.
func (p *pen) across(y, x0, x1 int, style diagram.LineStyle, r diagram.Role) {
	cx0, cy0 := p.at(x0, y)
	cx1, cy1 := p.at(x1, y)
	if p.l.orient.sideways {
		p.c.VLine(cx0, cy0, cy1, style, r)
	} else {
		p.c.HLine(cx0, cx1, cy0, style, r)
	}
}

// head draws a head at (x, y) of the layout, pointing in the direction d.
func (p *pen) head(x, y int, h Head, d dir) {
	cx, cy := p.at(x, y)
	p.c.Set(cx, cy, headGlyphs[p.turn(d)][h], diagram.RoleEdge)
}

// box draws the box of a node with its label.
func (p *pen) box(v vertex, sh Shape) {
	x0, y0, x1, y1 := p.canvasBox(v.x, v.y, v.w, v.h)
	drawBox(p.c, x0, y0, x1-x0+1, y1-y0+1, sh, v.lines)
}

// canvasBox returns the columns and rows on the canvas of the box at (x, y)
// of the layout, w by h.
func (p *pen) canvasBox(x, y, w, h int) (x0, y0, x1, y1 int) {
	x0, y0 = p.at(x, y)
	x1, y1 = p.at(x+w-1, y+h-1)
	return min(x0, x1), min(y0, y1), max(x0, x1), max(y0, y1)
}

// frame draws the outline of a frame.
func (p *pen) frame(f frame) {
	x0, y0, x1, y1 := p.canvasBox(f.x, f.y, f.w, f.h)
	for _, y := range []int{y0, y1} {
		p.c.HLine(x0, x1, y, diagram.LineSolid, diagram.RoleBorder)
	}
	for _, x := range []int{x0, x1} {
		p.c.VLine(x, y0, y1, diagram.LineSolid, diagram.RoleBorder)
	}
}

// title writes the title of a frame in its top on the canvas, after a
// corner and a line, as "┌─ title ─┐". It moves right past the edges that
// cross the top where it can, so that it doesn't hide them.
func (p *pen) title(f frame) {
	if f.title == "" {
		return
	}
	x0, y0, x1, _ := p.canvasBox(f.x, f.y, f.w, f.h)
	text := " " + diagram.FitLabel(f.title, x1-x0+1-6) + " "
	w := ansi.StringWidth(text)
	start := x0 + 2
	for x := start; x+w <= x1-1; x++ {
		if p.clear(x, y0, w) {
			start = x
			break
		}
	}
	p.c.Text(start, y0, text, diagram.RoleText)
}

// clear reports whether no lines cross row y from column x for w columns.
func (p *pen) clear(x, y, w int) bool {
	for i := range w {
		if p.c.Crossed(x+i, y) {
			return false
		}
	}
	return true
}

// text writes s from (x, y) of the layout to the right, for graphs that
// flow up or down.
func (p *pen) text(x, y int, s string) {
	cx, cy := p.at(x, y)
	p.c.Text(cx, cy, s, diagram.RoleEdgeLabel)
}

// textAlong writes s along column x of the layout, in the rows from y to y
// plus its width, for graphs that flow sideways, whose edges run along
// the canvas's rows there.
func (p *pen) textAlong(x, y int, s string) {
	cx0, cy := p.at(x, y)
	cx1, _ := p.at(x, y+ansi.StringWidth(s)-1)
	p.c.Text(min(cx0, cx1), cy, s, diagram.RoleEdgeLabel)
}
