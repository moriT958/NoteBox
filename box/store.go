package box

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Note struct {
	Title     string
	Size      int64
	CreatedAt time.Time
}

type IBox interface {
	Save(Note) error
	FindByTitle(string) (Note, error)
	FindAll() ([]Note, error)
	DeleteByTitle(string) error
}

var _ IBox = (*NoteBox)(nil)

type NoteBox struct {
	NoteNum     int
	storagePath string
}

func NewNoteBox(sp string) *NoteBox {
	nb := new(NoteBox)
	nb.NoteNum = 0
	nb.storagePath = sp
	return nb
}

func (b *NoteBox) Save(note Note) error {
	path := filepath.Join(b.storagePath, note.Title+".md")
	if note.Title == "" {
		return errors.New("failed to create note: title required")
	}

	if _, err := os.Stat(path); err == nil {
		return errors.New("this title already used")
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# %s", note.Title)
	fmt.Fprintln(os.Stdout, note.Title+" created!")

	return nil
}

func (b *NoteBox) FindByTitle(title string) (Note, error) {
	return Note{}, nil
}

func (b *NoteBox) FindAll() ([]Note, error) {
	notes := make([]Note, 0, b.NoteNum)
	filepath.Walk(b.storagePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == b.storagePath {
			return nil
		}

		n := Note{
			Title:     strings.TrimSuffix(info.Name(), filepath.Ext(path)),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		}
		notes = append(notes, n)

		return nil
	})

	return notes, nil
}

func (b *NoteBox) DeleteByTitle(title string) error {
	return nil
}
