package mermaid

import (
	"strings"
	"testing"
)

func lines(s ...string) string { return strings.Join(s, "\n") }

func TestCanvasLines(t *testing.T) {
	tests := []struct {
		name string
		draw func(c *canvas)
		want string
	}{
		{
			name: "box",
			draw: func(c *canvas) {
				c.hline(0, 4, 0, lineSolid, roleBorder)
				c.hline(0, 4, 2, lineSolid, roleBorder)
				c.vline(0, 0, 2, lineSolid, roleBorder)
				c.vline(4, 0, 2, lineSolid, roleBorder)
			},
			want: lines("┌───┐", "│   │", "└───┘"),
		},
		{
			name: "crossing",
			draw: func(c *canvas) {
				c.hline(0, 4, 1, lineSolid, roleEdge)
				c.vline(2, 0, 2, lineSolid, roleEdge)
			},
			want: lines("  │", "──┼──", "  │"),
		},
		{
			name: "junctions",
			draw: func(c *canvas) {
				// A box with an edge leaving its bottom side and one joining
				// its right side.
				c.hline(0, 4, 0, lineSolid, roleBorder)
				c.hline(0, 4, 2, lineSolid, roleBorder)
				c.vline(0, 0, 2, lineSolid, roleBorder)
				c.vline(4, 0, 2, lineSolid, roleBorder)
				c.vline(2, 2, 3, lineSolid, roleEdge)
				c.hline(4, 6, 1, lineSolid, roleEdge)
			},
			want: lines("┌───┐", "│   ├──", "└─┬─┘", "  │"),
		},
		{
			name: "elbow",
			draw: func(c *canvas) {
				c.vline(0, 0, 1, lineSolid, roleEdge)
				c.hline(0, 3, 1, lineSolid, roleEdge)
				c.vline(3, 1, 2, lineSolid, roleEdge)
			},
			want: lines("│", "└──┐", "   │"),
		},
		{
			name: "thick",
			draw: func(c *canvas) {
				c.vline(0, 0, 1, lineThick, roleEdge)
				c.hline(0, 2, 1, lineThick, roleEdge)
			},
			want: lines("┃", "┗━━"),
		},
		{
			name: "dotted lines have solid corners",
			draw: func(c *canvas) {
				c.vline(0, 0, 1, lineDotted, roleEdge)
				c.hline(0, 2, 1, lineDotted, roleEdge)
			},
			want: lines("┆", "└┄┄"),
		},
		{
			name: "mixed styles join solid",
			draw: func(c *canvas) {
				c.hline(0, 2, 1, lineThick, roleEdge)
				c.vline(1, 0, 2, lineDotted, roleEdge)
			},
			want: lines(" ┆", "━┼━", " ┆"),
		},
		{
			name: "single cell lines",
			draw: func(c *canvas) {
				c.hline(0, 0, 0, lineSolid, roleEdge)
				c.vline(2, 0, 0, lineSolid, roleEdge)
			},
			want: lines("─ │"),
		},
		{
			name: "lines keep text and glyphs",
			draw: func(c *canvas) {
				c.text(1, 0, "ab", roleText)
				c.set(4, 0, "▼", roleEdge)
				c.hline(0, 5, 0, lineSolid, roleEdge)
			},
			want: lines("─ab─▼─"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCanvas(7, 4)
			tt.draw(c)
			// Rows the drawing doesn't reach come out empty.
			if got := strings.TrimRight(c.String(), "\n"); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestCanvasText(t *testing.T) {
	tests := []struct {
		name      string
		draw      func(c *canvas) int
		want      string
		wantWidth int
	}{
		{
			name:      "wide graphemes",
			draw:      func(c *canvas) int { return c.text(0, 0, "aノートb", roleText) },
			want:      "aノートb",
			wantWidth: 8,
		},
		{
			name:      "cut at the edge",
			draw:      func(c *canvas) int { return c.text(5, 0, "ノート", roleText) },
			want:      "     ノ",
			wantWidth: 2,
		},
		{
			name:      "combining marks and control characters",
			draw:      func(c *canvas) int { return c.text(0, 0, "é\x1b[31m\tx", roleText) },
			want:      "é[31mx",
			wantWidth: 6,
		},
		{
			name: "overwriting the right half of a wide grapheme",
			draw: func(c *canvas) int {
				c.text(0, 0, "ノート", roleText)
				return c.text(1, 0, "x", roleText)
			},
			want:      " xート",
			wantWidth: 1,
		},
		{
			name: "overwriting the left half of a wide grapheme",
			draw: func(c *canvas) int {
				c.text(0, 0, "ノート", roleText)
				return c.text(2, 0, "x", roleText)
			},
			want:      "ノx ト",
			wantWidth: 1,
		},
		{
			name: "glyph over a wide grapheme",
			draw: func(c *canvas) int {
				c.text(0, 0, "ノ", roleText)
				c.set(1, 0, "▼", roleEdge)
				return 0
			},
			want: " ▼",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCanvas(8, 1)
			width := tt.draw(c)
			if got := c.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if width != tt.wantWidth {
				t.Errorf("width = %d, want %d", width, tt.wantWidth)
			}
		})
	}
}

func TestCanvasOutOfBounds(t *testing.T) {
	c := newCanvas(3, 2)
	c.hline(-2, 5, 0, lineSolid, roleEdge)
	c.vline(1, -1, 4, lineSolid, roleEdge)
	c.set(-1, 0, "x", roleEdge)
	c.set(3, 1, "x", roleEdge)
	c.text(-1, 1, "ab", roleText)
	if got, want := c.String(), lines("─┼─", " │"); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
