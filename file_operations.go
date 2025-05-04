package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

/* Note Entity */

type note struct {
	title string
	path  string
}

const (
	baseDir       string = "./notes"
	dummyNotePath string = "./dummy.md"
	defaultEditor string = "nvim"
)

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

func createNewNoteFileCmd(title string) tea.Cmd {
	timeStr := time.Now().Format(time.DateOnly)

	return func() tea.Msg {
		// TODO:
		// replace spaces with hyphen
		filename := filepath.Join(baseDir,
			title+"-"+timeStr+".md")

		fp, err := os.Create(filename)
		if err != nil {
			return errMsg{err}
		}
		defer fp.Close()

		content := fmt.Sprintf("# %s\n\n", title)
		fmt.Fprint(fp, content)

		note := note{
			title: title,
			path:  filename,
		}

		return createNewNoteMsg{note: note}
	}
}

func deleteNoteFile(title string) tea.Cmd {
	return func() tea.Msg {
		err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
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
	err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), title+"-") && strings.HasSuffix(info.Name(), ".md") {
			filename = filepath.Join(baseDir, info.Name())
			return io.EOF
		}

		return nil
	})
	if err != nil && err != io.EOF {
		return errCmd(err)
	}
	c := exec.Command(defaultEditor, filename)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return nil
	})
}
