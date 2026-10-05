package diagram

import (
	"strings"
	"testing"
)

func lines(s ...string) string { return strings.Join(s, "\n") }

func TestCanvasLines(t *testing.T) {
	tests := []struct {
		name string
		draw func(c *Canvas)
		want string
	}{
		{
			name: "box",
			draw: func(c *Canvas) {
				c.HLine(0, 4, 0, LineSolid, RoleBorder)
				c.HLine(0, 4, 2, LineSolid, RoleBorder)
				c.VLine(0, 0, 2, LineSolid, RoleBorder)
				c.VLine(4, 0, 2, LineSolid, RoleBorder)
			},
			want: lines("┌───┐", "│   │", "└───┘"),
		},
		{
			name: "crossing",
			draw: func(c *Canvas) {
				c.HLine(0, 4, 1, LineSolid, RoleEdge)
				c.VLine(2, 0, 2, LineSolid, RoleEdge)
			},
			want: lines("  │", "──┼──", "  │"),
		},
		{
			name: "junctions",
			draw: func(c *Canvas) {
				// A box with an edge leaving its bottom side and one joining
				// its right side.
				c.HLine(0, 4, 0, LineSolid, RoleBorder)
				c.HLine(0, 4, 2, LineSolid, RoleBorder)
				c.VLine(0, 0, 2, LineSolid, RoleBorder)
				c.VLine(4, 0, 2, LineSolid, RoleBorder)
				c.VLine(2, 2, 3, LineSolid, RoleEdge)
				c.HLine(4, 6, 1, LineSolid, RoleEdge)
			},
			want: lines("┌───┐", "│   ├──", "└─┬─┘", "  │"),
		},
		{
			name: "elbow",
			draw: func(c *Canvas) {
				c.VLine(0, 0, 1, LineSolid, RoleEdge)
				c.HLine(0, 3, 1, LineSolid, RoleEdge)
				c.VLine(3, 1, 2, LineSolid, RoleEdge)
			},
			want: lines("│", "└──┐", "   │"),
		},
		{
			name: "thick",
			draw: func(c *Canvas) {
				c.VLine(0, 0, 1, LineThick, RoleEdge)
				c.HLine(0, 2, 1, LineThick, RoleEdge)
			},
			want: lines("┃", "┗━━"),
		},
		{
			name: "dotted lines have solid corners",
			draw: func(c *Canvas) {
				c.VLine(0, 0, 1, LineDotted, RoleEdge)
				c.HLine(0, 2, 1, LineDotted, RoleEdge)
			},
			want: lines("┆", "└┄┄"),
		},
		{
			name: "mixed styles join solid",
			draw: func(c *Canvas) {
				c.HLine(0, 2, 1, LineThick, RoleEdge)
				c.VLine(1, 0, 2, LineDotted, RoleEdge)
			},
			want: lines(" ┆", "━┼━", " ┆"),
		},
		{
			name: "single cell lines",
			draw: func(c *Canvas) {
				c.HLine(0, 0, 0, LineSolid, RoleEdge)
				c.VLine(2, 0, 0, LineSolid, RoleEdge)
			},
			want: lines("─ │"),
		},
		{
			name: "lines keep text and glyphs",
			draw: func(c *Canvas) {
				c.Text(1, 0, "ab", RoleText)
				c.Set(4, 0, "▼", RoleEdge)
				c.HLine(0, 5, 0, LineSolid, RoleEdge)
			},
			want: lines("─ab─▼─"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(7, 4)
			tt.draw(c)
			// Rows the drawing doesn't reach come out empty.
			if got := strings.TrimRight(c.Render(Styles{}), "\n"); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestCanvasText(t *testing.T) {
	tests := []struct {
		name      string
		draw      func(c *Canvas) int
		want      string
		wantWidth int
	}{
		{
			name:      "wide graphemes",
			draw:      func(c *Canvas) int { return c.Text(0, 0, "aノートb", RoleText) },
			want:      "aノートb",
			wantWidth: 8,
		},
		{
			name:      "cut at the edge",
			draw:      func(c *Canvas) int { return c.Text(5, 0, "ノート", RoleText) },
			want:      "     ノ",
			wantWidth: 2,
		},
		{
			name:      "combining marks and control characters",
			draw:      func(c *Canvas) int { return c.Text(0, 0, "é\x1b[31m\tx", RoleText) },
			want:      "é[31mx",
			wantWidth: 6,
		},
		{
			name: "overwriting the right half of a wide grapheme",
			draw: func(c *Canvas) int {
				c.Text(0, 0, "ノート", RoleText)
				return c.Text(1, 0, "x", RoleText)
			},
			want:      " xート",
			wantWidth: 1,
		},
		{
			name: "overwriting the left half of a wide grapheme",
			draw: func(c *Canvas) int {
				c.Text(0, 0, "ノート", RoleText)
				return c.Text(2, 0, "x", RoleText)
			},
			want:      "ノx ト",
			wantWidth: 1,
		},
		{
			name: "glyph over a wide grapheme",
			draw: func(c *Canvas) int {
				c.Text(0, 0, "ノ", RoleText)
				c.Set(1, 0, "▼", RoleEdge)
				return 0
			},
			want: " ▼",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(8, 1)
			width := tt.draw(c)
			if got := c.Render(Styles{}); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if width != tt.wantWidth {
				t.Errorf("width = %d, want %d", width, tt.wantWidth)
			}
		})
	}
}

func TestCanvasOutOfBounds(t *testing.T) {
	c := New(3, 2)
	c.HLine(-2, 5, 0, LineSolid, RoleEdge)
	c.VLine(1, -1, 4, LineSolid, RoleEdge)
	c.Set(-1, 0, "x", RoleEdge)
	c.Set(3, 1, "x", RoleEdge)
	c.Text(-1, 1, "ab", RoleText)
	if got, want := c.Render(Styles{}), lines("─┼─", " │"); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
