package model

type Note struct {
	ID      string
	Title   string
	Content string
}

type INoteStore interface {
	Save(Note) error
	GetAll() ([]Note, error)
	GetByID(id string) (Note, error)
	DeleteByID(id string) error
}
