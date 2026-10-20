package preview

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSelectedText(t *testing.T) {
	tests := []struct {
		name       string
		lines      []string
		start, end pos
		want       string
	}{
		{
			name:  "Within a line, both ends are included.",
			lines: []string{"  hello world"},
			start: pos{0, 2}, end: pos{0, 6},
			want: "hello",
		},
		{
			name:  "Colors are dropped.",
			lines: []string{"  \x1b[1mbold\x1b[0m text"},
			start: pos{0, 0}, end: pos{0, 12},
			want: "bold text",
		},
		{
			name:  "Trailing spaces are dropped.",
			lines: []string{"  one      ", "  two      "},
			start: pos{0, 0}, end: pos{1, 10},
			want: "one\ntwo",
		},
		{
			name:  "The shared margin is dropped, deeper indentation is kept.",
			lines: []string{"  func f() {", "      return", "", "  }"},
			start: pos{0, 0}, end: pos{3, 10},
			want: "func f() {\n    return\n\n}",
		},
		{
			name:  "A first line that starts mid-line doesn't keep the margin of the rest.",
			lines: []string{"  first line", "  second line"},
			start: pos{0, 8}, end: pos{1, 20},
			want: "line\nsecond line",
		},
		{
			name:  "A wide character cut in half at either end is kept whole.",
			lines: []string{"あいうえ"},
			start: pos{0, 1}, end: pos{0, 4},
			want: "あいう",
		},
		{
			name:  "Lines past the end of the note are ignored.",
			lines: []string{"  last"},
			start: pos{0, 0}, end: pos{5, 3},
			want: "last",
		},
		{
			name:  "A selection below the note is empty.",
			lines: []string{"  last"},
			start: pos{2, 0}, end: pos{3, 3},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectedText(tt.lines, tt.start, tt.end); got != tt.want {
				t.Errorf("selectedText() = %q, want %q", got, tt.want)
			}
		})
	}
}

// numberedLines returns n lines "  line a", "  line b", ...
func numberedLines(n int) string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = "  line " + string(rune('a'+i))
	}
	return strings.Join(lines, "\n")
}

func TestPreview_Selection(t *testing.T) {
	t.Run("Dragging selects from the press to the release, in either direction.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 40)
		p.SetRendered(notes[0], numberedLines(5), false)

		// Dragged upward: from the end of "line c" to the start of "line b".
		p.StartSelection(7, 2)
		p.ExtendSelection(2, 1)
		text, ok := p.EndSelection()

		if !ok || text != "line b\nline c" {
			t.Errorf("EndSelection() = %q, %v, want %q, true", text, ok, "line b\nline c")
		}
	})

	t.Run("The selection stays on the same text while the note scrolls.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 40)
		p.SetRendered(notes[0], numberedLines(20), false)

		p.StartSelection(2, 0)
		p.Wheel(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
		p.ExtendSelection(7, 0)
		text, _ := p.EndSelection()

		// The wheel scrolls 3 lines, so the pointer ends on the 4th line.
		if want := "line a\nline b\nline c\nline d"; text != want {
			t.Errorf("EndSelection() = %q, want %q", text, want)
		}
	})

	t.Run("A drag out of the body is held to its edges.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 40)
		p.SetRendered(notes[0], numberedLines(20), false)

		p.StartSelection(-5, -5)
		p.ExtendSelection(100, 1)
		text, _ := p.EndSelection()

		if want := "line a\nline b"; text != want {
			t.Errorf("EndSelection() = %q, want %q", text, want)
		}
	})

	t.Run("A click without a drag selects nothing.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 40)
		p.SetRendered(notes[0], numberedLines(5), false)

		p.StartSelection(3, 1)
		_, ok := p.EndSelection()

		if ok || p.sel != nil {
			t.Errorf("expected no selection after a click")
		}
	})

	t.Run("The selection is highlighted until the content changes.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 40)
		p.SetRendered(notes[0], numberedLines(5), false)
		plain := p.Render(true)

		p.StartSelection(2, 0)
		p.ExtendSelection(7, 0)
		p.EndSelection()
		if p.Render(true) == plain {
			t.Errorf("expected the selection to be highlighted")
		}

		p.SetRendered(notes[0], numberedLines(6), false)
		if p.sel != nil {
			t.Errorf("expected a re-render to clear the selection")
		}
	})
}
