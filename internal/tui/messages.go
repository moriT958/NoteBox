package tui

import (
	"fmt"
	"notebox/internal/note"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
)

type errMsg error

func errCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg(err)
	}
}

// this contains content of the note.
type renderPreviewMsg string

func (m model) renderPreviewCmd(path string) tea.Cmd {
	return func() tea.Msg {
		var content string
		if len(m.listPanel.items) > 0 {
			b, err := os.ReadFile(path)
			if err != nil {
				return errMsg(err)
			}
			content = string(b)
		} else {
			b, err := os.ReadFile(m.cfg.DummyNoteDir)
			if err != nil {
				return errMsg(err)
			}
			content = string(b)
		}
		return renderPreviewMsg(string(content))
	}
}

// this inform tea of if note file succesessfull created.
type newNoteCreatedMsg note.Note

func createNewNoteCmd(notesdir, title string) tea.Cmd {
	createdTime := time.Now().Format(time.DateOnly)
	return func() tea.Msg {
		filename := filepath.Join(notesdir, title+"-"+createdTime+".md")
		fp, err := os.Create(filename)
		if err != nil {
			return errMsg(err)
		}
		defer fp.Close()

		content := fmt.Sprintf("# %s\n\n", title)
		fmt.Fprint(fp, content)

		newNote := note.Note{
			Title: title,
			Path:  filename,
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
