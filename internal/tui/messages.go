package tui

import (
	"context"
	"notebox/internal/config"
	"notebox/internal/notedeprecated"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type errMsg error

func errCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg(err)
	}
}

// this inform tea of if note file succesessfull created.
type newNoteCreatedMsg notedeprecated.Note

func createNewNoteCmd(notesdir, title string) tea.Cmd {
	return func() tea.Msg {
		newNote, err := notedeprecated.CreateNote(notesdir, title)
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

func renameNoteCmd(n notedeprecated.Note, newTitle string) tea.Cmd {
	return func() tea.Msg {
		_, err := notedeprecated.RenameNote(n, newTitle)
		if err != nil {
			return errMsg(err)
		}
		return nil
	}
}

type notesChangedMsg []notedeprecated.Note

func waitNoteChangeCmd(ch <-chan []notedeprecated.Note) tea.Cmd {
	return func() tea.Msg {
		notes, ok := <-ch
		if !ok {
			return nil
		}
		return notesChangedMsg(notes)
	}
}

type boxesLoadedMsg []notedeprecated.Box

func loadBoxesCmd(repo notedeprecated.BoxRepository) tea.Cmd {
	return func() tea.Msg {
		boxes, err := repo.FindAllActive(context.Background())
		if err != nil {
			return errMsg(err)
		}
		return boxesLoadedMsg(boxes)
	}
}

type boxCreatedMsg notedeprecated.Box

// resolveBoxPath expands ~, and resolves relative paths against cwd.
func resolveBoxPath(input, cwd, home string) string {
	if input == "~" {
		return home
	}
	if strings.HasPrefix(input, "~/") {
		return filepath.Join(home, input[2:])
	}
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	return filepath.Join(cwd, input)
}

// newBoxFinalPath returns the path for a new box directory.
// If resolvedBase is empty the box is placed under configDir.
func newBoxFinalPath(title, resolvedBase, configDir string) string {
	if resolvedBase == "" {
		return filepath.Join(configDir, title)
	}
	return filepath.Join(resolvedBase, title)
}

// isDuplicatePath reports whether path is already used by one of boxes.
func isDuplicatePath(path string, boxes []notedeprecated.Box) bool {
	for _, b := range boxes {
		if b.Path == path {
			return true
		}
	}
	return false
}

// newBoxCmd creates the directory finalPath (if needed) and registers it as a box.
func newBoxCmd(repo notedeprecated.BoxRepository, title, finalPath string) tea.Cmd {
	return func() tea.Msg {
		if err := os.MkdirAll(finalPath, 0755); err != nil {
			return errMsg(err)
		}
		box, err := repo.CreateBox(context.Background(), notedeprecated.Box{Title: title, Path: finalPath})
		if err != nil {
			return errMsg(err)
		}
		return boxCreatedMsg(box)
	}
}

// openFolderAsBoxCmd registers an existing directory as a box without modifying the filesystem.
func openFolderAsBoxCmd(repo notedeprecated.BoxRepository, title, path string) tea.Cmd {
	return func() tea.Msg {
		box, err := repo.CreateBox(context.Background(), notedeprecated.Box{Title: title, Path: path})
		if err != nil {
			return errMsg(err)
		}
		return boxCreatedMsg(box)
	}
}

type boxDeletedMsg int

func deleteBoxCmd(repo notedeprecated.BoxRepository, box notedeprecated.Box) tea.Cmd {
	return func() tea.Msg {
		if err := repo.DeleteBox(context.Background(), box); err != nil {
			return errMsg(err)
		}
		return boxDeletedMsg(box.ID)
	}
}

type boxRenamedMsg notedeprecated.Box

func renameBoxCmd(repo notedeprecated.BoxRepository, box notedeprecated.Box) tea.Cmd {
	return func() tea.Msg {
		updated, err := repo.UpdateBox(context.Background(), box)
		if err != nil {
			return errMsg(err)
		}
		return boxRenamedMsg(updated)
	}
}

func saveLastBoxCmd(id int) tea.Cmd {
	return func() tea.Msg {
		_ = config.SaveLastBoxID(id)
		return nil
	}
}
