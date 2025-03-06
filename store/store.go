package store

type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type Store interface {
	Save(Note) (int, error)
	FindByID(int) (Note, error)
	FindAll() ([]Note, error)
	DeleteByID(int) error
}

type NoteStore struct {
	DB string
}

func NewNoteStore(db string) *NoteStore {
	return &NoteStore{
		DB: db,
	}
}

var _ Store = (*NoteStore)(nil)

func (s *NoteStore) Save(note Note) (int, error) {
	return 0, nil
}

func (s *NoteStore) FindByID(id int) (Note, error) {
	return Note{}, nil
}

func (s *NoteStore) FindAll() ([]Note, error) {
	return nil, nil
}

func (s *NoteStore) DeleteByID(id int) error {
	return nil
}
