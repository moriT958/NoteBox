package tui

import (
	"notebox/internal/note"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

type errMsg error

func errCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg(err)
	}
}

// renderPreviewMsg contains rendered content of the note.
type renderPreviewMsg string

func renderPreviewCmd(renderer note.NoteRenderer, n note.Note) tea.Cmd {
	return func() tea.Msg {
		rendered, err := renderer.RenderNote(n)
		if err != nil {
			return errMsg(err)
		}
		return renderPreviewMsg(rendered)
	}
}

// this inform tea of if note file succesessfull created.
type newNoteCreatedMsg note.Note

func createNewNoteCmd(notesdir, title string) tea.Cmd {
	return func() tea.Msg {
		newNote, err := note.CreateNote(notesdir, title)
		if err != nil {
			return errMsg(err)
		}
		return newNoteCreatedMsg(newNote)
	}
}

func openNoteWithEditor(editor, path string) tea.Cmd {
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg(err)
		}
		return nil
	})
}

func deleteNoteFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		if err := os.Remove(path); err != nil {
			return errCmd(err)
		}
		return nil
	}
}

func renameNoteCmd(n note.Note, newTitle string) tea.Cmd {
	return func() tea.Msg {
		_, err := note.RenameNote(n, newTitle)
		if err != nil {
			return errMsg(err)
		}
		return nil
	}
}

type notesChangedMsg []note.Note

func waitNoteChangeCmd(ch <-chan []note.Note) tea.Cmd {
	return func() tea.Msg {
		notes, ok := <-ch
		if !ok {
			return nil
		}
		return notesChangedMsg(notes)
	}
}
