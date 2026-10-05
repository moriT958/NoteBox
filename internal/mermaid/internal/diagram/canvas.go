// Package diagram holds what drawing any kind of diagram needs: a canvas of
// terminal cells, the text of labels, colors and the errors diagrams that
// aren't drawn are reported with.
package diagram

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// A Role tells which part of a diagram a cell belongs to, which decides its
// color.
type Role uint8

const (
	roleNone Role = iota
	RoleBorder
	RoleText
	RoleEdge
	RoleEdgeLabel
)

// A LineStyle is how a line is drawn: solid, dotted or thick.
type LineStyle uint8

const (
	LineSolid LineStyle = iota
	LineDotted
	LineThick
)

type cell struct {
	// s is the grapheme drawn in the cell. It is empty in the cell to the
	// right of a wide grapheme, which that grapheme covers.
	s    string
	role Role
	// dirs and style describe the line drawn through the cell, if any. A
	// cell with lines is redrawn as more lines join it.
	dirs  dirs
	style LineStyle
}

// A Canvas is a grid of terminal cells that diagrams are drawn on. Lines
// drawn across each other join into one box-drawing character, and wide
// graphemes keep whole.
type Canvas struct {
	w, h  int
	cells []cell
}

// New returns an empty canvas of w columns and h rows.
func New(w, h int) *Canvas {
	c := &Canvas{w: w, h: h, cells: make([]cell, w*h)}
	for i := range c.cells {
		c.cells[i].s = " "
	}
	return c
}

func (c *Canvas) at(x, y int) *cell {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return nil
	}
	return &c.cells[y*c.w+x]
}

// put draws a grapheme of width w at (x, y), first clearing any wide
// grapheme it would overlap.
func (c *Canvas) put(x, y int, s string, w int, r Role) {
	for i := range w {
		cl := c.at(x+i, y)
		if cl.s == "" {
			// The right half of a wide grapheme starting further left.
			if left := c.at(x+i-1, y); left != nil {
				*left = cell{s: " "}
			}
		} else if ansi.StringWidth(cl.s) == 2 {
			if right := c.at(x+i+1, y); right != nil {
				*right = cell{s: " "}
			}
		}
	}
	*c.at(x, y) = cell{s: s, role: r}
	for i := 1; i < w; i++ {
		*c.at(x+i, y) = cell{role: r}
	}
}

// Set draws a single-width glyph at (x, y), such as an arrowhead or a box
// corner. Lines drawn later don't replace it.
func (c *Canvas) Set(x, y int, glyph string, r Role) {
	if c.at(x, y) == nil {
		return
	}
	c.put(x, y, glyph, 1, r)
}

// Text draws s from (x, y) to the right and returns its width. Graphemes
// that don't fit in the canvas, and ones without width such as control
// characters, are dropped.
func (c *Canvas) Text(x, y int, s string, r Role) int {
	start := x
	for s != "" {
		g, w := ansi.FirstGraphemeCluster(s, ansi.GraphemeWidth)
		s = s[len(g):]
		if w == 0 || strings.ContainsFunc(g, isControl) {
			continue
		}
		if c.at(x, y) == nil || c.at(x+w-1, y) == nil {
			break
		}
		c.put(x, y, g, w, r)
		x += w
	}
	return x - start
}

func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0)
}

// line adds the directions d to the line through (x, y) and redraws the cell
// with the box-drawing character joining them all. Cells with text keep it.
func (c *Canvas) line(x, y int, d dirs, style LineStyle, r Role) {
	cl := c.at(x, y)
	if cl == nil || d == 0 {
		return
	}
	if cl.dirs == 0 {
		if cl.s != " " {
			return
		}
		cl.style = style
	} else if cl.style != style {
		cl.style = LineSolid
	}
	cl.dirs |= d
	cl.role = r
	cl.s = lineGlyph(cl.dirs, cl.style)
}

// HLine draws a horizontal line from x0 to x1 on row y.
func (c *Canvas) HLine(x0, x1, y int, style LineStyle, r Role) {
	x0, x1 = min(x0, x1), max(x0, x1)
	for x := x0; x <= x1; x++ {
		var d dirs
		if x > x0 {
			d |= dirLeft
		}
		if x < x1 {
			d |= dirRight
		}
		if x0 == x1 {
			d = dirLeft | dirRight
		}
		c.line(x, y, d, style, r)
	}
}

// VLine draws a vertical line from y0 to y1 in column x.
func (c *Canvas) VLine(x, y0, y1 int, style LineStyle, r Role) {
	y0, y1 = min(y0, y1), max(y0, y1)
	for y := y0; y <= y1; y++ {
		var d dirs
		if y > y0 {
			d |= dirUp
		}
		if y < y1 {
			d |= dirDown
		}
		if y0 == y1 {
			d = dirUp | dirDown
		}
		c.line(x, y, d, style, r)
	}
}
