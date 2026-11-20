package note

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrConflict is returned by WriteNoteContent when the note changed since the
// version the caller based its edit on.
var ErrConflict = errors.New("note was changed by someone else")

// ETag identifies a version of a note's content. It is derived from the bytes
// only, so the same content always has the same ETag no matter who wrote it.
func ETag(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:8])
}

// AbsPath returns the absolute path of the note's file.
func (s *NoteService) AbsPath(n Note) string {
	return filepath.Join(s.box.Path, n.path)
}

// ReadNote returns the content of the note and its ETag.
func (s *NoteService) ReadNote(n Note) ([]byte, string, error) {
	content, err := os.ReadFile(s.AbsPath(n))
	if err != nil {
		return nil, "", fmt.Errorf("failed to read note: %w", err)
	}
	return content, ETag(content), nil
}

// WriteNoteContent replaces the content of the note. When ifMatch is not
// empty, the write only happens if the current content still has that ETag;
// otherwise ErrConflict is returned along with the current ETag.
//
// The file is replaced atomically (temp file + rename) so that readers never
// see a half-written note.
func (s *NoteService) WriteNoteContent(n Note, content []byte, ifMatch string) (string, error) {
	path := s.AbsPath(n)
	if ifMatch != "" {
		cur, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to write note: %w", err)
		}
		if tag := ETag(cur); tag != ifMatch {
			return tag, ErrConflict
		}
	}

	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".nbx-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to write note: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return "", fmt.Errorf("failed to write note: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("failed to write note: %w", err)
	}
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return "", fmt.Errorf("failed to write note: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", fmt.Errorf("failed to write note: %w", err)
	}
	return ETag(content), nil
}
