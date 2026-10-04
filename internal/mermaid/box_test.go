package mermaid

import "testing"

func TestDrawBox(t *testing.T) {
	tests := []struct {
		name  string
		w, h  int
		shape shape
		lines []string
		want  string
	}{
		{
			name:  "rectangle",
			w:     7,
			h:     3,
			shape: shapeRect,
			lines: []string{"abc"},
			want:  lines("┌─────┐", "│ abc │", "└─────┘"),
		},
		{
			name:  "rounded",
			w:     7,
			h:     3,
			shape: shapeRound,
			lines: []string{"abc"},
			want:  lines("╭─────╮", "│ abc │", "╰─────╯"),
		},
		{
			name:  "diamond",
			w:     7,
			h:     3,
			shape: shapeDiamond,
			lines: []string{"abc"},
			want:  lines("╱─────╲", "│ abc │", "╲─────╱"),
		},
		{
			name:  "lines centered",
			w:     10,
			h:     4,
			shape: shapeRect,
			lines: []string{"ノート", "a"},
			want:  lines("┌────────┐", "│ ノート │", "│   a    │", "└────────┘"),
		},
		{
			name:  "too wide lines cut short",
			w:     7,
			h:     3,
			shape: shapeRect,
			lines: []string{"abcdefg"},
			want:  lines("┌─────┐", "│abcd…│", "└─────┘"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCanvas(tt.w, tt.h)
			drawBox(c, 0, 0, tt.w, tt.h, tt.shape, tt.lines)
			if got := c.String(); got != tt.want {
				t.Errorf("got\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestDrawBoxRoles(t *testing.T) {
	c := newCanvas(5, 3)
	drawBox(c, 0, 0, 5, 3, shapeRect, []string{"a"})
	for _, tt := range []struct {
		x, y int
		want role
	}{
		{0, 0, roleBorder},
		{2, 0, roleBorder},
		{0, 1, roleBorder},
		{2, 1, roleText},
		{4, 2, roleBorder},
	} {
		if got := c.at(tt.x, tt.y).role; got != tt.want {
			t.Errorf("role at (%d, %d) = %d, want %d", tt.x, tt.y, got, tt.want)
		}
	}
}
