package note

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type NoteService struct {
	box noteBox
}

type noteBox interface {
	Path() string
}

func NewNoteService(box noteBox) *NoteService {
	return &NoteService{
		box: box,
	}
}

func (s *NoteService) GetNotes() ([]Note, error) {
	notes := make([]Note, 0)

	if err := filepath.Walk(s.box.Path(), func(path string, info fs.FileInfo, err error) error {
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

		title := strings.TrimSuffix(filename, ".md")
		note := &Note{
			title: title,
			path:  path,
		}
		notes = append(notes, *note)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	return notes, nil
}

func (s *NoteService) GetNoteContent(note Note) (string, error) {
	notePath := filepath.Join(s.box.Path(), note.path)
	content, err := os.ReadFile(notePath)
	if err != nil {
		return "", fmt.Errorf("failed to get note content: %w", err)
	}

	return string(content), nil
}

func (s *NoteService) CreateNote(title string) (Note, error) {
	relPath := title + ".md"
	absPath := filepath.Join(s.box.Path(), relPath)

	fp, err := os.Create(absPath)
	if err != nil {
		return Note{}, fmt.Errorf("failed to create note: %w", err)
	}
	defer fp.Close()

	content := fmt.Sprintf("# %s\n\n", title)
	fmt.Fprint(fp, content)

	return Note{
		title: title,
		path:  relPath,
	}, nil
}

func (s *NoteService) RenameNote(note Note, newTitle string) (Note, error) {
	boxPath := filepath.Dir(s.box.Path())
	oldNotePath := filepath.Join(boxPath, note.path)
	newNotePath := filepath.Join(boxPath, newTitle+".md")

	if err := os.Rename(oldNotePath, newNotePath); err != nil {
		return Note{}, fmt.Errorf("failed to rename note: %w", err)
	}

	return Note{
		title: newTitle,
		path:  newNotePath,
	}, nil
}

func (s *NoteService) RemoveNote(note Note) error {
	notePath := filepath.Join(s.box.Path(), note.path)
	if err := os.Remove(notePath); err != nil {
		return fmt.Errorf("failed to remove note: %w", err)
	}
	return nil
}
