package graph

import (
	"notebox/internal/mermaid/internal/diagram"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDrawBox(t *testing.T) {
	tests := []struct {
		name  string
		w, h  int
		shape Shape
		lines []string
		want  string
	}{
		{
			name:  "rectangle",
			w:     7,
			h:     3,
			shape: ShapeRect,
			lines: []string{"abc"},
			want:  lines("┌─────┐", "│ abc │", "└─────┘"),
		},
		{
			name:  "rounded",
			w:     7,
			h:     3,
			shape: ShapeRound,
			lines: []string{"abc"},
			want:  lines("╭─────╮", "│ abc │", "╰─────╯"),
		},
		{
			name:  "diamond",
			w:     7,
			h:     3,
			shape: ShapeDiamond,
			lines: []string{"abc"},
			want:  lines("╱─────╲", "│ abc │", "╲─────╱"),
		},
		{
			name:  "lines centered",
			w:     10,
			h:     4,
			shape: ShapeRect,
			lines: []string{"ノート", "a"},
			want:  lines("┌────────┐", "│ ノート │", "│   a    │", "└────────┘"),
		},
		{
			name:  "too wide lines cut short",
			w:     7,
			h:     3,
			shape: ShapeRect,
			lines: []string{"abcdefg"},
			want:  lines("┌─────┐", "│abcd…│", "└─────┘"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := diagram.New(tt.w, tt.h)
			drawBox(c, 0, 0, tt.w, tt.h, tt.shape, tt.lines)
			if got := c.Render(diagram.Styles{}); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestDrawBoxRoles(t *testing.T) {
	c := diagram.New(5, 3)
	drawBox(c, 0, 0, 5, 3, ShapeRect, []string{"a"})
	got := c.Render(diagram.Styles{Border: ansi.IndexedColor(1), Text: ansi.IndexedColor(2)})
	const border, text, reset = "\x1b[38;5;1m", "\x1b[38;5;2m", "\x1b[39m"
	want := lines(
		border+"┌───┐"+reset,
		border+"│ "+reset+text+"a "+reset+border+"│"+reset,
		border+"└───┘"+reset,
	)
	if got != want {
		t.Errorf("got\n%q\nwant\n%q", got, want)
	}
}

func lines(s ...string) string { return strings.Join(s, "\n") }
