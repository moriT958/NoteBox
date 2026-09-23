package note

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"notebox/internal/core/box"
)

type NoteService struct {
	box box.Box
}

func NewNoteService(b box.Box) *NoteService {
	return &NoteService{
		box: b,
	}
}

func (s *NoteService) GetNotes() ([]Note, error) {
	notes := make([]Note, 0)

	if err := filepath.Walk(s.box.Path, func(path string, info fs.FileInfo, err error) error {
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

		relPath, err := filepath.Rel(s.box.Path, path)
		if err != nil {
			return err
		}

		title := decodeTitle(strings.TrimSuffix(filename, ".md"))
		note := &Note{
			title: title,
			path:  relPath,
		}
		notes = append(notes, *note)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	return notes, nil
}

func (s *NoteService) GetNoteContent(note Note) (string, error) {
	notePath := filepath.Join(s.box.Path, note.path)
	content, err := os.ReadFile(notePath)
	if err != nil {
		return "", fmt.Errorf("failed to get note content: %w", err)
	}

	return string(content), nil
}

func (s *NoteService) CreateNote(title string) (Note, error) {
	relPath := encodeTitle(title) + ".md"
	absPath := filepath.Join(s.box.Path, relPath)

	fp, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return Note{}, fmt.Errorf("failed to create note: note with this title already exists")
		}
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
	dir := filepath.Dir(note.path)
	newRelPath := filepath.Join(dir, encodeTitle(newTitle)+".md")

	oldNotePath := filepath.Join(s.box.Path, note.path)
	newNotePath := filepath.Join(s.box.Path, newRelPath)

	if newNotePath != oldNotePath {
		if _, err := os.Stat(newNotePath); err == nil {
			return Note{}, fmt.Errorf("failed to rename note: note with this title already exists")
		} else if !errors.Is(err, fs.ErrNotExist) {
			return Note{}, fmt.Errorf("failed to rename note: %w", err)
		}
	}

	if err := os.Rename(oldNotePath, newNotePath); err != nil {
		return Note{}, fmt.Errorf("failed to rename note: %w", err)
	}

	return Note{
		title: newTitle,
		path:  newRelPath,
	}, nil
}

func (s *NoteService) RemoveNote(note Note) error {
	notePath := filepath.Join(s.box.Path, note.path)
	if err := os.Remove(notePath); err != nil {
		return fmt.Errorf("failed to remove note: %w", err)
	}
	return nil
}
