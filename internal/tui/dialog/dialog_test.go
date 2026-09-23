package dialog

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"notebox/internal/core/box"
	"notebox/internal/core/note"
	"notebox/internal/tui/common"
	"notebox/internal/tui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func newCommon(t *testing.T) *common.Common {
	t.Helper()
	sty, err := styles.New("dark")
	if err != nil {
		t.Fatalf("failed to create styles: %v", err)
	}
	return &common.Common{Styles: sty}
}

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

func press(k string) tea.KeyPressMsg {
	switch k {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	r := []rune(k)[0]
	return tea.KeyPressMsg{Code: r, Text: k}
}

func typeText(d Dialog, s string) {
	for _, r := range s {
		d.HandleMsg(press(string(r)))
	}
}

func TestOverlay(t *testing.T) {
	com := newCommon(t)
	o := NewOverlay()
	o.Open(NewBoxes(com, ""))
	o.Open(NewConfirm(com, "sure?", ActionClose{}))

	if got := o.Front().ID(); got != ConfirmID {
		t.Errorf("front = %q, want %q", got, ConfirmID)
	}

	o.CloseFront()
	if got := o.Front().ID(); got != BoxesID {
		t.Errorf("front = %q, want %q", got, BoxesID)
	}

	o.Close(BoxesID)
	if o.HasDialogs() {
		t.Errorf("expected no dialogs")
	}
}

func TestInput(t *testing.T) {
	com := newCommon(t)

	t.Run("Submits the typed title.", func(t *testing.T) {
		d := NewNoteInput(com)
		typeText(d, "hello")

		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionCreateNote); !ok || a.Title != "hello" {
			t.Errorf("action = %#v, want ActionCreateNote{hello}", got)
		}
	})

	t.Run("Ignores enter on an empty title.", func(t *testing.T) {
		d := NewNoteInput(com)
		if got := d.HandleMsg(press("enter")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}
	})
}

func TestConfirm(t *testing.T) {
	com := newCommon(t)
	yes := ActionDeleteBox{Box: box.Box{ID: "1"}}

	if got := NewConfirm(com, "sure?", yes).HandleMsg(press("enter")); got != yes {
		t.Errorf("enter: action = %#v, want %#v", got, yes)
	}
	for _, k := range []string{"esc", "q"} {
		if got := NewConfirm(com, "sure?", yes).HandleMsg(press(k)); got != (ActionClose{}) {
			t.Errorf("%s: action = %#v, want ActionClose", k, got)
		}
	}
}

func TestFinder(t *testing.T) {
	com := newCommon(t)
	notes := newNotes(t, "apple", "banana", "cherry")

	t.Run("Selects the fuzzy match.", func(t *testing.T) {
		d := NewFinder(com, notes)
		typeText(d, "bna")

		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionSelectNote); !ok || a.Note.Title() != "banana" {
			t.Errorf("action = %#v, want ActionSelectNote{banana}", got)
		}
	})

	t.Run("Moves through all notes without a query.", func(t *testing.T) {
		d := NewFinder(com, notes)
		d.HandleMsg(press("down"))
		d.HandleMsg(press("down"))
		d.HandleMsg(press("down"))

		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionSelectNote); !ok || a.Note.Title() != "cherry" {
			t.Errorf("action = %#v, want ActionSelectNote{cherry}", got)
		}
	})

	t.Run("Does nothing on enter without matches.", func(t *testing.T) {
		d := NewFinder(com, notes)
		typeText(d, "zzz")
		if got := d.HandleMsg(press("enter")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}
	})
}

func TestBoxes(t *testing.T) {
	com := newCommon(t)
	current := box.Box{ID: "1", Title: "Current"}
	other := box.Box{ID: "2", Title: "Other"}
	items := []BoxItem{{Box: current}, {Box: other}}

	t.Run("Ignores selection while loading.", func(t *testing.T) {
		d := NewBoxes(com, current.ID)
		if got := d.HandleMsg(press("enter")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}
	})

	t.Run("Starts on the given box and switches to the selected one.", func(t *testing.T) {
		d := NewBoxes(com, current.ID)
		d.SetBoxes(items, other.ID)

		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionSwitchBox); !ok || a.Box.ID != other.ID {
			t.Errorf("action = %#v, want ActionSwitchBox{2}", got)
		}
	})

	t.Run("Doesn't delete the current box.", func(t *testing.T) {
		d := NewBoxes(com, current.ID)
		d.SetBoxes(items, current.ID)
		if got := d.HandleMsg(press("d")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}
	})

	t.Run("Asks before deleting another box.", func(t *testing.T) {
		d := NewBoxes(com, current.ID)
		d.SetBoxes(items, other.ID)

		open, ok := d.HandleMsg(press("d")).(ActionOpen)
		if !ok {
			t.Fatalf("expected a confirm dialog to open")
		}
		got := open.Dialog.HandleMsg(press("enter"))

		if a, ok := got.(ActionDeleteBox); !ok || a.Box.ID != other.ID {
			t.Errorf("action = %#v, want ActionDeleteBox{2}", got)
		}
	})

	t.Run("Renames the selected box inline.", func(t *testing.T) {
		d := NewBoxes(com, current.ID)
		d.SetBoxes(items, other.ID)

		d.HandleMsg(press("r"))
		typeText(d, "!")
		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionRenameBox); !ok || a.Box.ID != other.ID || a.Title != "Other!" {
			t.Errorf("action = %#v, want ActionRenameBox{2, Other!}", got)
		}
		if d.renaming {
			t.Errorf("expected rename to stop")
		}
	})

	t.Run("Esc while renaming only stops renaming.", func(t *testing.T) {
		d := NewBoxes(com, current.ID)
		d.SetBoxes(items, other.ID)
		d.HandleMsg(press("r"))

		if got := d.HandleMsg(press("esc")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}
		if d.renaming {
			t.Errorf("expected rename to stop")
		}
	})
}

func TestBoxForm(t *testing.T) {
	com := newCommon(t)

	t.Run("Requires a name.", func(t *testing.T) {
		d := NewBoxForm(com, BoxFormNew)
		if got := d.HandleMsg(press("enter")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}
		if d.err == "" {
			t.Errorf("expected an error to be shown")
		}
	})

	t.Run("Creates a box with an optional path.", func(t *testing.T) {
		d := NewBoxForm(com, BoxFormNew)
		typeText(d, "work")

		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionCreateBox); !ok || a.Title != "work" || a.Path != "" {
			t.Errorf("action = %#v, want ActionCreateBox{work, \"\"}", got)
		}
	})

	t.Run("Opening a folder requires a path.", func(t *testing.T) {
		d := NewBoxForm(com, BoxFormOpenFolder)
		typeText(d, "work")
		if got := d.HandleMsg(press("enter")); got != nil {
			t.Errorf("action = %#v, want nil", got)
		}

		d.HandleMsg(press("tab"))
		typeText(d, "~/work")
		got := d.HandleMsg(press("enter"))

		if a, ok := got.(ActionOpenFolder); !ok || a.Title != "work" || a.Path != "~/work" {
			t.Errorf("action = %#v, want ActionOpenFolder{work, ~/work}", got)
		}
	})
}

func TestRender_FitsSmallArea(t *testing.T) {
	com := newCommon(t)
	notes := newNotes(t, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l")
	boxes := NewBoxes(com, "")
	items := make([]BoxItem, 20)
	for i := range items {
		items[i] = BoxItem{Box: box.Box{ID: string(rune('a' + i)), Title: "a box with a rather long title"}}
	}
	boxes.SetBoxes(items, "")

	dialogs := []Dialog{
		NewNoteInput(com),
		NewConfirm(com, "sure?", ActionClose{}),
		NewFinder(com, notes),
		boxes,
		NewBoxForm(com, BoxFormNew),
	}
	area := image.Rect(0, 0, 40, 16)
	for _, d := range dialogs {
		view, _ := d.Render(area)
		if w := lipgloss.Width(view); w > area.Dx() {
			t.Errorf("%s: width = %d, want <= %d", d.ID(), w, area.Dx())
		}
		if h := lipgloss.Height(view); h > area.Dy() {
			t.Errorf("%s: height = %d, want <= %d", d.ID(), h, area.Dy())
		}
	}
}
