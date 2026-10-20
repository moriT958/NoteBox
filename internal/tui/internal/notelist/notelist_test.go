package notelist

import (
	"os"
	"path/filepath"
	"testing"

	"notebox/internal/core/box"
	"notebox/internal/core/note"
	"notebox/internal/tui/internal/common"
	"notebox/internal/tui/styles"
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

// newList returns a list tall enough to show visible notes at once.
func newList(t *testing.T, visible int) *NoteList {
	t.Helper()
	sty := styles.New(styles.DarkPalette())
	l := New(&common.Common{Styles: sty})
	l.SetSize(30, visible*itemLines+sty.BorderFocused.GetVerticalFrameSize())
	return l
}

func selectedTitle(t *testing.T, l *NoteList) string {
	t.Helper()
	n, ok := l.Selected()
	if !ok {
		t.Fatalf("expected a selected note")
	}
	return n.Title()
}

func TestNoteList_SetNotes(t *testing.T) {
	t.Run("Keeps the selected note selected when it still exists.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c")
		l := newList(t, 5)
		l.SetNotes(notes)
		l.CursorDown() // b

		l.SetNotes([]note.Note{notes[0], notes[2], notes[1]})

		if got := selectedTitle(t, l); got != "b" {
			t.Errorf("selected = %q, want %q", got, "b")
		}
	})

	t.Run("Keeps the cursor index when the selected note is gone.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c")
		l := newList(t, 5)
		l.SetNotes(notes)
		l.CursorDown() // b

		l.SetNotes([]note.Note{notes[0], notes[2]})

		if got := selectedTitle(t, l); got != "c" {
			t.Errorf("selected = %q, want %q", got, "c")
		}
	})

	t.Run("Clamps the cursor when the list shrinks below it.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c")
		l := newList(t, 5)
		l.SetNotes(notes)
		l.CursorDown()
		l.CursorDown() // c

		l.SetNotes(notes[:1])

		if got := selectedTitle(t, l); got != "a" {
			t.Errorf("selected = %q, want %q", got, "a")
		}
	})
}

func TestNoteList_Remove(t *testing.T) {
	t.Run("Selects the next note after removing the selected one.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c")
		l := newList(t, 5)
		l.SetNotes(notes)
		l.CursorDown() // b

		l.Remove(notes[1])

		if len(l.Notes()) != 2 {
			t.Errorf("len(notes) = %d, want 2", len(l.Notes()))
		}
		if got := selectedTitle(t, l); got != "c" {
			t.Errorf("selected = %q, want %q", got, "c")
		}
	})
}

func TestNoteList_Scroll(t *testing.T) {
	t.Run("Scrolls to keep the cursor visible.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c", "d", "e")
		l := newList(t, 2)
		l.SetNotes(notes)

		for range 3 {
			l.CursorDown()
		}

		if got := selectedTitle(t, l); got != "d" {
			t.Errorf("selected = %q, want %q", got, "d")
		}
		if l.offset != 2 {
			t.Errorf("offset = %d, want 2", l.offset)
		}

		for range 3 {
			l.CursorUp()
		}
		if l.offset != 0 {
			t.Errorf("offset = %d, want 0", l.offset)
		}
	})

	t.Run("Keeps the cursor visible after the list gets shorter.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c", "d", "e")
		l := newList(t, 2)
		l.SetNotes(notes)
		for range 4 {
			l.CursorDown()
		}

		l.SetSize(30, 1*itemLines+2)

		if l.cursor < l.offset || l.cursor >= l.offset+l.visibleItems() {
			t.Errorf("cursor %d not visible from offset %d", l.cursor, l.offset)
		}
	})
}

func TestNoteList_Rename(t *testing.T) {
	t.Run("Starts with the selected note's title.", func(t *testing.T) {
		l := newList(t, 5)
		l.SetNotes(newNotes(t, "a"))

		if !l.StartRename() {
			t.Fatalf("expected rename to start")
		}
		if got := l.RenameValue(); got != "a" {
			t.Errorf("rename value = %q, want %q", got, "a")
		}
		if l.Cursor() == nil {
			t.Errorf("expected a cursor while renaming")
		}

		l.StopRename()
		if l.Renaming() || l.Cursor() != nil {
			t.Errorf("expected rename to stop")
		}
	})

	t.Run("Doesn't start on an empty list.", func(t *testing.T) {
		l := newList(t, 5)
		if l.StartRename() {
			t.Errorf("expected rename not to start")
		}
	})
}

func TestNoteList_SelectAt(t *testing.T) {
	t.Run("Selects the note whose title or path is at the row, below the scrolled-off notes.", func(t *testing.T) {
		notes := newNotes(t, "a", "b", "c", "d", "e")
		l := newList(t, 2)
		l.SetNotes(notes)
		for range 3 {
			l.CursorDown() // d, scrolled to show c and d
		}

		for row, want := range map[int]string{0: "c", 1: "c", 3: "d", 4: "d"} {
			if !l.SelectAt(row) {
				t.Fatalf("SelectAt(%d) = false, want true", row)
			}
			if got := selectedTitle(t, l); got != want {
				t.Errorf("SelectAt(%d) selected %q, want %q", row, got, want)
			}
		}
	})

	t.Run("Selects nothing at a spacer, below the last note, or out of the body.", func(t *testing.T) {
		notes := newNotes(t, "a", "b")
		l := newList(t, 5)
		l.SetNotes(notes)

		for _, row := range []int{2, 5, 6, -1, 100} {
			if l.SelectAt(row) {
				t.Errorf("SelectAt(%d) = true, want false", row)
			}
			if got := selectedTitle(t, l); got != "a" {
				t.Errorf("SelectAt(%d) moved the cursor to %q", row, got)
			}
		}
	})
}

func TestNoteList_Hover(t *testing.T) {
	t.Run("Marks the note under the pointer, but not the one under the cursor.", func(t *testing.T) {
		notes := newNotes(t, "a", "b")
		l := newList(t, 5)
		l.SetNotes(notes)
		plain := l.Render(true)

		l.SetHover(0) // a, under the cursor
		if l.Render(true) != plain {
			t.Errorf("expected no mark on the note under the cursor")
		}

		l.SetHover(4) // b's path
		if l.Render(true) == plain {
			t.Errorf("expected the note under the pointer to be marked")
		}

		l.ClearHover()
		if l.Render(true) != plain {
			t.Errorf("expected no mark after the pointer left")
		}
	})
}
