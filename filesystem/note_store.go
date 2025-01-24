package filesystem

import (
	"fmt"
	"notebox/model"
	"os"
	"sync"
)

var (
	Store *NoteStore
	once  sync.Once
)

var _ model.INoteStore = (*NoteStore)(nil)

type NoteStore struct{}

func init() {
	once.Do(func() {
		Store = new(NoteStore)
	})
}

func (r *NoteStore) Save(note model.Note) error {
	filename := note.ID + ".md"
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, fmt.Sprintf("# %s", note.Title))

	return nil
}

func (r *NoteStore) GetAll() ([]model.Note, error) {
	return nil, nil
}

func (r *NoteStore) GetByID(id string) (model.Note, error) {
	return model.Note{}, nil
}

func (r *NoteStore) DeleteByID(id string) error {
	return nil
}
