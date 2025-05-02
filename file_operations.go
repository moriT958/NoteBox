package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

/* Note Entity */

type note struct {
	title   string
	content string
}

const baseDir string = "./notes"

func loadNoteFiles(baseDir string) ([]note, error) {
	notes := make([]note, 0)

	if err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, filename := filepath.Split(path)
		title := getTitleFromFilename(filename)
		notes = append(notes, note{title, string(content)})

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

func createNewNoteFile(title string) tea.Cmd {
	timeStr := time.Now().Format(time.DateOnly)

	return func() tea.Msg {
		filename := filepath.Join(baseDir,
			title+"-"+timeStr+".md")

		fp, err := os.Create(filename)
		if err != nil {
			return errMsg{err}
		}
		defer fp.Close()

		content := fmt.Sprintf("# %s\n\n", title)
		fmt.Fprint(fp, content)

		return nil
	}
}
