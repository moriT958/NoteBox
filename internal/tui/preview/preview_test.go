package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"notebox/internal/core/box"
	"notebox/internal/core/note"
	"notebox/internal/tui/common"
	"notebox/internal/tui/styles"

	"github.com/charmbracelet/x/ansi"
)

// newNotes creates a note file per title and returns them in list order.
func newNotes(t *testing.T, titles ...string) []note.Note {
	t.Helper()
	dir := t.TempDir()
	for _, title := range titles {
		if err := os.WriteFile(filepath.Join(dir, title+".md"), nil, 0644); err != nil {
			t.Fatalf("failed to create note: %v", err)
		}
	}
	notes, err := note.NewNoteService(box.Box{Path: dir}).GetNotes()
	if err != nil {
		t.Fatalf("failed to get notes: %v", err)
	}
	return notes
}

func newPreview(t *testing.T, width int) *Preview {
	t.Helper()
	sty, err := styles.New("dark")
	if err != nil {
		t.Fatalf("failed to create styles: %v", err)
	}
	p := New(&common.Common{Styles: sty})
	p.SetSize(width, 10)
	return p
}

func activeTitle(t *testing.T, p *Preview) string {
	t.Helper()
	n, ok := p.Active()
	if !ok {
		t.Fatalf("expected an active tab")
	}
	return n.Title()
}

func TestPreview_SetRendered(t *testing.T) {
	t.Run("A preview tab is replaced by the next preview.", func(t *testing.T) {
		notes := newNotes(t, "a", "b")
		p := newPreview(t, 80)

		p.SetRendered(notes[0], "A", false)
		p.SetRendered(notes[1], "B", false)

		if len(p.Notes()) != 1 {
			t.Fatalf("len(tabs) = %d, want 1", len(p.Notes()))
		}
		if got := activeTitle(t, p); got != "b" {
			t.Errorf("active = %q, want %q", got, "b")
		}
	})

	t.Run("A pinned tab stays when another note is previewed.", func(t *testing.T) {
		notes := newNotes(t, "a", "b")
		p := newPreview(t, 80)

		p.SetRendered(notes[0], "A", true)
		p.SetRendered(notes[1], "B", false)

		if len(p.Notes()) != 2 {
			t.Fatalf("len(tabs) = %d, want 2", len(p.Notes()))
		}
		if got := activeTitle(t, p); got != "b" {
			t.Errorf("active = %q, want %q", got, "b")
		}
	})

	t.Run("Re-rendering an open tab keeps its place and pin state.", func(t *testing.T) {
		notes := newNotes(t, "a", "b")
		p := newPreview(t, 80)
		p.SetRendered(notes[0], "A", true)
		p.SetRendered(notes[1], "B", false)

		p.SetRendered(notes[0], "A2", false)

		if got := activeTitle(t, p); got != "b" {
			t.Errorf("active = %q, want %q", got, "b")
		}
		if p.tabs[0].preview {
			t.Errorf("tab a should stay pinned")
		}
		if p.tabs[0].rendered != "A2" {
			t.Errorf("rendered = %q, want %q", p.tabs[0].rendered, "A2")
		}
	})
}

func TestPreview_Open(t *testing.T) {
	t.Run("Reports whether the note needs rendering.", func(t *testing.T) {
		notes := newNotes(t, "a", "b")
		p := newPreview(t, 80)
		p.SetRendered(notes[0], "A", false)

		if !p.Open(notes[0], false) {
			t.Errorf("Open(a) = false, want true")
		}
		if p.Open(notes[1], false) {
			t.Errorf("Open(b) = true, want false")
		}
	})

	t.Run("Pins an open preview tab.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 80)
		p.SetRendered(notes[0], "A", false)

		p.Open(notes[0], true)

		if p.tabs[0].preview {
			t.Errorf("tab should be pinned")
		}
	})
}

func TestPreview_Close(t *testing.T) {
	t.Run("Keeps at least one tab open.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 80)
		p.SetRendered(notes[0], "A", true)

		p.CloseActive()

		if len(p.Notes()) != 1 {
			t.Errorf("len(tabs) = %d, want 1", len(p.Notes()))
		}
	})

	t.Run("Activates the neighbor after removing the active tab.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c")
		p := newPreview(t, 80)
		for _, n := range notes {
			p.SetRendered(n, n.Title(), true)
		}
		p.PrevTab() // b

		p.Remove(notes[1])

		if got := activeTitle(t, p); got != "c" {
			t.Errorf("active = %q, want %q", got, "c")
		}
	})

	t.Run("Keeps the active tab when removing one before it.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c")
		p := newPreview(t, 80)
		for _, n := range notes {
			p.SetRendered(n, n.Title(), true)
		}

		p.Remove(notes[0])

		if got := activeTitle(t, p); got != "c" {
			t.Errorf("active = %q, want %q", got, "c")
		}
	})

	t.Run("Removing the last tab empties the preview.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 80)
		p.SetRendered(notes[0], "A", false)

		p.Remove(notes[0])

		if _, ok := p.Active(); ok {
			t.Errorf("expected no active tab")
		}
	})
}

func TestPreview_Placeholder(t *testing.T) {
	t.Run("Shows the placeholder only while no tab is open.", func(t *testing.T) {
		notes := newNotes(t, "a")
		p := newPreview(t, 80)
		if !strings.Contains(p.Render(false), "No Note Selected") {
			t.Errorf("expected the placeholder in an empty preview")
		}

		p.SetRendered(notes[0], "A", false)
		if strings.Contains(p.Render(false), "No Note Selected") {
			t.Errorf("expected no placeholder with a tab open")
		}

		p.Remove(notes[0])
		if !strings.Contains(p.Render(false), "No Note Selected") {
			t.Errorf("expected the placeholder after the last tab is closed")
		}
	})
}

func TestPreview_TabBar(t *testing.T) {
	sty, err := styles.New("dark")
	if err != nil {
		t.Fatalf("failed to create styles: %v", err)
	}

	titles := make([]string, 12)
	for i := range titles {
		titles[i] = fmt.Sprintf("note %02d with a long title", i)
	}
	notes := newNotes(t, titles...)

	for _, width := range []int{2, 5, 7, 20, 21, 33, 45, 80} {
		for count := 0; count <= len(notes); count++ {
			t.Run(fmt.Sprintf("width %d, %d tabs", width, count), func(t *testing.T) {
				p := newPreview(t, width)
				for _, n := range notes[:count] {
					p.SetRendered(n, "", true)
				}

				for step := 0; step <= count; step++ {
					bar := ansi.Strip(p.renderTabBar(sty.FrameFocused, sty.Tab.Focused))
					if got := ansi.StringWidth(bar); got != width {
						t.Fatalf("tab bar width = %d, want %d: %q", got, width, bar)
					}
					if count > 0 && len(p.tabWidths(p.offset)) > 0 {
						visible := len(p.tabWidths(p.offset))
						if p.active < p.offset || p.active >= p.offset+visible {
							t.Fatalf("active %d not within visible tabs [%d, %d)", p.active, p.offset, p.offset+visible)
						}
					}
					p.PrevTab()
				}
			})
		}
	}
}
