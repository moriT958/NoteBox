package diagram

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCanvasRender(t *testing.T) {
	const (
		border = "\x1b[38;5;1m"
		text   = "\x1b[38;5;2m"
		edge   = "\x1b[38;5;3m"
		label  = "\x1b[38;5;4m"
		reset  = "\x1b[39m"
	)
	styles := Styles{
		Border:    ansi.IndexedColor(1),
		Text:      ansi.IndexedColor(2),
		Edge:      ansi.IndexedColor(3),
		EdgeLabel: ansi.IndexedColor(4),
	}

	tests := []struct {
		name   string
		draw   func(c *Canvas)
		styles Styles
		want   string
	}{
		{
			name: "runs switch colors and spaces keep the current one",
			draw: func(c *Canvas) {
				c.HLine(0, 3, 0, LineSolid, RoleBorder)
				c.Text(5, 0, "ab", RoleText)
				c.HLine(7, 8, 0, LineSolid, RoleEdge)
				c.Text(1, 1, "ノ x", RoleEdgeLabel)
			},
			styles: styles,
			want: lines(
				border+"──── "+reset+text+"ab"+reset+edge+"──"+reset,
				" "+label+"ノ x"+reset,
			),
		},
		{
			name: "trailing spaces are trimmed outside the colors",
			draw: func(c *Canvas) {
				c.Text(0, 0, "a", RoleText)
			},
			styles: styles,
			want:   lines(text+"a"+reset, ""),
		},
		{
			name: "parts without a color reset to the default",
			draw: func(c *Canvas) {
				c.HLine(0, 1, 0, LineSolid, RoleBorder)
				c.Text(2, 0, "a", RoleText)
				c.HLine(3, 4, 0, LineSolid, RoleBorder)
			},
			styles: Styles{Border: ansi.IndexedColor(1)},
			want:   lines(border+"──"+reset+"a"+border+"──"+reset, ""),
		},
		{
			name: "no colors",
			draw: func(c *Canvas) {
				c.HLine(0, 1, 0, LineSolid, RoleBorder)
				c.Text(2, 0, "a", RoleText)
			},
			want: lines("──a", ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(10, 2)
			tt.draw(c)
			if got := c.Render(tt.styles); got != tt.want {
				t.Errorf("got\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
