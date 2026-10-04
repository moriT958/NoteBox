package mermaid

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// A role tells which part of a diagram a cell belongs to.
type role uint8

const (
	roleNone role = iota
	roleBorder
	roleText
	roleEdge
	roleEdgeLabel
)

// dirs is a set of directions in which a line leaves a cell.
type dirs uint8

const (
	dirUp dirs = 1 << iota
	dirDown
	dirLeft
	dirRight
)

type lineStyle uint8

const (
	lineSolid lineStyle = iota
	lineDotted
	lineThick
)

type cell struct {
	// s is the grapheme drawn in the cell. It is empty in the cell to the
	// right of a wide grapheme, which that grapheme covers.
	s    string
	role role
	// dirs and style describe the line drawn through the cell, if any. A
	// cell with lines is redrawn as more lines join it.
	dirs  dirs
	style lineStyle
}

// A canvas is a grid of terminal cells that diagrams are drawn on.
type canvas struct {
	w, h  int
	cells []cell
}

func newCanvas(w, h int) *canvas {
	c := &canvas{w: w, h: h, cells: make([]cell, w*h)}
	for i := range c.cells {
		c.cells[i].s = " "
	}
	return c
}

func (c *canvas) at(x, y int) *cell {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		return nil
	}
	return &c.cells[y*c.w+x]
}

// put draws a grapheme of width w at (x, y), first clearing any wide
// grapheme it would overlap.
func (c *canvas) put(x, y int, s string, w int, r role) {
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

// set draws a single-width glyph at (x, y), such as an arrowhead or a box
// corner. Lines drawn later don't replace it.
func (c *canvas) set(x, y int, glyph string, r role) {
	if c.at(x, y) == nil {
		return
	}
	c.put(x, y, glyph, 1, r)
}

// text draws s from (x, y) to the right and returns its width. Graphemes
// that don't fit in the canvas, and ones without width such as control
// characters, are dropped.
func (c *canvas) text(x, y int, s string, r role) int {
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
func (c *canvas) line(x, y int, d dirs, style lineStyle, r role) {
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
		cl.style = lineSolid
	}
	cl.dirs |= d
	cl.role = r
	cl.s = lineGlyph(cl.dirs, cl.style)
}

// hline draws a horizontal line from x0 to x1 on row y.
func (c *canvas) hline(x0, x1, y int, style lineStyle, r role) {
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

// vline draws a vertical line from y0 to y1 in column x.
func (c *canvas) vline(x, y0, y1 int, style lineStyle, r role) {
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

var solidGlyphs = map[dirs]string{
	dirUp:                                "│",
	dirDown:                              "│",
	dirUp | dirDown:                      "│",
	dirLeft:                              "─",
	dirRight:                             "─",
	dirLeft | dirRight:                   "─",
	dirDown | dirRight:                   "┌",
	dirDown | dirLeft:                    "┐",
	dirUp | dirRight:                     "└",
	dirUp | dirLeft:                      "┘",
	dirUp | dirDown | dirRight:           "├",
	dirUp | dirDown | dirLeft:            "┤",
	dirDown | dirLeft | dirRight:         "┬",
	dirUp | dirLeft | dirRight:           "┴",
	dirUp | dirDown | dirLeft | dirRight: "┼",
}

var thickGlyphs = map[dirs]string{
	dirUp:                                "┃",
	dirDown:                              "┃",
	dirUp | dirDown:                      "┃",
	dirLeft:                              "━",
	dirRight:                             "━",
	dirLeft | dirRight:                   "━",
	dirDown | dirRight:                   "┏",
	dirDown | dirLeft:                    "┓",
	dirUp | dirRight:                     "┗",
	dirUp | dirLeft:                      "┛",
	dirUp | dirDown | dirRight:           "┣",
	dirUp | dirDown | dirLeft:            "┫",
	dirDown | dirLeft | dirRight:         "┳",
	dirUp | dirLeft | dirRight:           "┻",
	dirUp | dirDown | dirLeft | dirRight: "╋",
}

// lineGlyph returns the box-drawing character for a line leaving a cell in
// the directions d. Dotted lines only have straight characters, so their
// corners and junctions are solid.
func lineGlyph(d dirs, style lineStyle) string {
	switch style {
	case lineThick:
		return thickGlyphs[d]
	case lineDotted:
		switch d {
		case dirUp, dirDown, dirUp | dirDown:
			return "┆"
		case dirLeft, dirRight, dirLeft | dirRight:
			return "┄"
		}
	}
	return solidGlyphs[d]
}

// String returns the canvas as lines of text without trailing spaces.
func (c *canvas) String() string {
	var b strings.Builder
	for y := range c.h {
		var row strings.Builder
		for x := range c.w {
			row.WriteString(c.cells[y*c.w+x].s)
		}
		b.WriteString(strings.TrimRight(row.String(), " "))
		if y < c.h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
