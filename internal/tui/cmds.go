package tui

import (
	"context"
	"os"
	"os/exec"

	"notebox/internal/config"
	"notebox/internal/core/box"
	"notebox/internal/core/note"
	"notebox/internal/tui/dialog"
	"notebox/internal/tui/preview"

	tea "charm.land/bubbletea/v2"
)

// Every command below does IO off the update loop and reports back with a
// message; none of them touches the UI's state, so services are passed in
// rather than read from the UI. Messages about notes carry the ID of the box
// they belong to, so results that arrive after switching boxes can be
// dropped.

type errMsg struct{ err error }

type (
	notesChangedMsg struct{ ch <-chan struct{} }
	notesLoadedMsg  struct {
		boxID string
		notes []note.Note
	}
	noteRenderedMsg struct {
		boxID    string
		note     note.Note
		rendered string
		pin      bool
	}
	noteCreatedMsg struct {
		boxID string
		note  note.Note
	}
	noteRenamedMsg struct {
		boxID    string
		old, new note.Note
	}
	noteRemovedMsg struct {
		boxID string
		note  note.Note
	}
)

type (
	boxesLoadedMsg struct {
		items    []dialog.BoxItem
		selectID string
	}
	boxCreatedMsg struct{ box box.Box }
	boxFormErrMsg struct{ err error }
	boxRenamedMsg struct{ box box.Box }
	boxRemovedMsg struct{}
)

// watchNotesCmd waits for the next change in the box directory.
func watchNotesCmd(ch <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		if _, ok := <-ch; !ok {
			return nil
		}
		return notesChangedMsg{ch: ch}
	}
}

func loadNotesCmd(svc *note.NoteService, boxID string) tea.Cmd {
	return func() tea.Msg {
		notes, err := svc.GetNotes()
		if err != nil {
			return errMsg{err}
		}
		return notesLoadedMsg{boxID: boxID, notes: notes}
	}
}

func renderNoteCmd(svc *note.NoteService, r *preview.Renderer, boxID string, n note.Note, pin bool) tea.Cmd {
	return func() tea.Msg {
		content, err := svc.GetNoteContent(n)
		if err != nil {
			return errMsg{err}
		}
		rendered, err := r.Render(content)
		if err != nil {
			return errMsg{err}
		}
		return noteRenderedMsg{boxID: boxID, note: n, rendered: rendered, pin: pin}
	}
}

func createNoteCmd(svc *note.NoteService, boxID, title string) tea.Cmd {
	return func() tea.Msg {
		n, err := svc.CreateNote(title)
		if err != nil {
			return errMsg{err}
		}
		return noteCreatedMsg{boxID: boxID, note: n}
	}
}

func renameNoteCmd(svc *note.NoteService, boxID string, n note.Note, title string) tea.Cmd {
	return func() tea.Msg {
		renamed, err := svc.RenameNote(n, title)
		if err != nil {
			return errMsg{err}
		}
		return noteRenamedMsg{boxID: boxID, old: n, new: renamed}
	}
}

func removeNoteCmd(svc *note.NoteService, boxID string, n note.Note) tea.Cmd {
	return func() tea.Msg {
		if err := svc.RemoveNote(n); err != nil {
			return errMsg{err}
		}
		return noteRemovedMsg{boxID: boxID, note: n}
	}
}

func editNoteCmd(editor, path string) tea.Cmd {
	return tea.ExecProcess(exec.Command(editor, path), func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return nil
	})
}

func loadBoxesCmd(svc *box.BoxService, selectID string) tea.Cmd {
	return func() tea.Msg {
		boxes, err := svc.GetActiveBoxes(context.Background())
		if err != nil {
			return errMsg{err}
		}
		items := make([]dialog.BoxItem, len(boxes))
		for i, b := range boxes {
			_, err := os.Stat(b.Path)
			items[i] = dialog.BoxItem{Box: b, Missing: os.IsNotExist(err)}
		}
		return boxesLoadedMsg{items: items, selectID: selectID}
	}
}

// createBoxCmd creates a box directory under base, or under the default
// location when base is empty.
func createBoxCmd(svc *box.BoxService, title, base string) tea.Cmd {
	return func() tea.Msg {
		var basePtr *string
		if base != "" {
			basePtr = &base
		}
		b, err := svc.CreateBox(context.Background(), title, basePtr)
		if err != nil {
			return boxFormErrMsg{err}
		}
		return boxCreatedMsg{box: *b}
	}
}

func openFolderCmd(svc *box.BoxService, title, path string) tea.Cmd {
	return func() tea.Msg {
		b, err := svc.OpenFolderAsBox(context.Background(), title, path)
		if err != nil {
			return boxFormErrMsg{err}
		}
		return boxCreatedMsg{box: *b}
	}
}

func renameBoxCmd(svc *box.BoxService, id, title string) tea.Cmd {
	return func() tea.Msg {
		b, err := svc.RenameBox(context.Background(), id, title)
		if err != nil {
			return errMsg{err}
		}
		return boxRenamedMsg{box: *b}
	}
}

func removeBoxCmd(svc *box.BoxService, id string) tea.Cmd {
	return func() tea.Msg {
		if err := svc.RemoveBox(context.Background(), id); err != nil {
			return errMsg{err}
		}
		return boxRemovedMsg{}
	}
}

func saveLastBoxCmd(id string) tea.Cmd {
	return func() tea.Msg {
		if err := config.SaveLastBoxID(id); err != nil {
			return errMsg{err}
		}
		return nil
	}
}
