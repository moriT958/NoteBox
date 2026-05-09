package note

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func LoadNoteFiles(notesDir string) ([]Note, error) {
	notes := make([]Note, 0)

	if err := filepath.Walk(notesDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		_, filename := filepath.Split(path)
		if filepath.Ext(filename) != ".md" {
			return nil
		}

		title := getTitleFromFilename(filename)
		note := &Note{
			Title: title,
			Path:  path,
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
	if len(parts) <= 3 {
		return filename
	}

	title := strings.Join(parts[:len(parts)-3], "-")
	return title
}

func CreateNote(notesDir, title string) (Note, error) {
	createdTime := time.Now().Format(time.DateOnly)

	filename := filepath.Join(notesDir, title+"-"+createdTime+".md")
	fp, err := os.Create(filename)
	if err != nil {
		return Note{}, err
	}
	defer fp.Close()

	content := fmt.Sprintf("# %s\n\n", title)
	fmt.Fprint(fp, content)

	return Note{
		Title: title,
		Path:  filename,
	}, nil
}

func RenameNote(note Note, newTitle string) (Note, error) {
	renamedNote := renameNote(note, newTitle)
	if err := os.Rename(note.Path, renamedNote.Path); err != nil {
		return Note{}, err
	}
	return renamedNote, nil
}

func renameNote(note Note, newTitle string) Note {
	dir := filepath.Dir(note.Path)
	stem := strings.TrimSuffix(filepath.Base(note.Path), ".md")
	parts := strings.Split(stem, "-")

	var newPath string
	if len(parts) > 3 {
		dateSuffix := strings.Join(parts[len(parts)-3:], "-")
		newPath = filepath.Join(dir, newTitle+"-"+dateSuffix+".md")
	} else {
		newPath = filepath.Join(dir, newTitle+"-"+time.Now().Format(time.DateOnly)+".md")
	}

	return Note{
		Title: newTitle,
		Path:  newPath,
	}
}
