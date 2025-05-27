package main

import (
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"notebox/config"
)

/* Note Entity */

type note struct {
	title string
	path  string
}

func loadNoteFiles(baseDir string) ([]note, error) {
	notes := make([]note, 0)

	if err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		_, filename := filepath.Split(path)
		title := getTitleFromFilename(filename)
		note := &note{
			title: title,
			path:  path,
		}
		notes = append(notes, *note)

		return nil
	}); err != nil {
		return nil, err
	}

	return notes, nil
}

func getTitleFromFilename(filename string) string {
	filename = strings.TrimSuffix(filename, ".md")
	parts := strings.Split(filename, "-")

	title := strings.Join(parts[:len(parts)-3], "-")
	return title
}

func deleteNoteFileCmd(title string) tea.Cmd {
	return func() tea.Msg {
		err := filepath.Walk(config.NotesDir(), func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			filename := info.Name()
			if strings.HasPrefix(filename, title+"-") && strings.HasSuffix(filename, ".md") {
				if err := os.Remove(path); err != nil {
					return err
				}
				return io.EOF
			}

			return nil
		})
		if err != nil && err != io.EOF {
			return errMsg{err}
		}

		return nil
	}
}

func openNoteWithEditor(title string) tea.Cmd {
	var filename string
	err := filepath.Walk(config.NotesDir(), func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), title+"-") && strings.HasSuffix(info.Name(), ".md") {
			filename = filepath.Join(config.NotesDir(), info.Name())
			return io.EOF
		}

		return nil
	})
	if err != nil && err != io.EOF {
		return errCmd(err)
	}
	c := exec.Command(config.Editor(), filename)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return nil
	})
}
