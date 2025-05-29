package tui

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func loadNoteFiles(notesDir string) ([]note, error) {
	notes := make([]note, 0)

	if err := filepath.Walk(notesDir, func(path string, info fs.FileInfo, err error) error {
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
