package store

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type Store interface {
	Save(Note) (int, error)
	FindByID(int) (Note, error)
	FindAll() []Note
	DeleteByID(int) error
}

type NoteStore struct {
	mem []Note
}

func NewNoteStore(dbFile string) (*NoteStore, error) {

	fp, err := os.Open(dbFile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	dbBytes, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	var data []Note
	if err := json.Unmarshal(dbBytes, &data); err != nil {
		return nil, err
	}

	return &NoteStore{mem: data}, nil
}

var _ Store = (*NoteStore)(nil)

func (s *NoteStore) Save(note Note) (int, error) {

	return 0, nil
}

func (s *NoteStore) FindByID(id int) (Note, error) {
	// リニアサーチ
	for i := range len(s.mem) {
		if s.mem[i].ID == id {
			return s.mem[i], nil
		}
	}
	return Note{}, errors.New("id doesn't match")
}

func (s *NoteStore) FindAll() []Note {
	return s.mem
}

func (s *NoteStore) DeleteByID(id int) error {
	return nil
}
