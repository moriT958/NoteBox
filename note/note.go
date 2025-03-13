package note

import "database/sql"

type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type Repository interface {
	Save(Note) (int, error)
	FindByID(int) (Note, error)
	FindAll() ([]Note, error)
	DeleteByID(int) error
}

type NoteRepository struct {
	DB *sql.DB
}

const dbSchema = `
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL,
	path TEXT NOT NULL
);
`

func NewNoteRepository(db *sql.DB) (*NoteRepository, error) {
	// Migrate database
	if _, err := db.Exec(dbSchema); err != nil {
		return nil, err
	}
	return &NoteRepository{DB: db}, nil
}

// Implement Repository interface
var _ Repository = (*NoteRepository)(nil)

func (r *NoteRepository) Save(note Note) (int, error) {

	if _, err := r.DB.Exec(`INSERT INTO notes (title, path) VALUES (?, ?);`, note.Title, note.Path); err != nil {
		return 0, err
	}

	return note.ID, nil
}

func (r *NoteRepository) FindByID(id int) (Note, error) {

	note := new(Note)

	if err := r.DB.QueryRow(`SELECT id, title, path FROM notes WHERE id = ?;`, id).
		Scan(&note.ID, &note.Title, &note.Path); err != nil {
		return Note{}, err
	}

	return *note, nil
}

func (r *NoteRepository) FindAll() ([]Note, error) {

	notes := make([]Note, 0)

	rows, err := r.DB.Query(`SELECT * FROM notes;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		n := new(Note)
		if err := rows.Scan(&n.ID, &n.Title, &n.Path); err != nil {
			return nil, err
		}
		notes = append(notes, *n)
	}

	return notes, nil
}

func (r *NoteRepository) DeleteByID(id int) error {

	if _, err := r.DB.Exec(`DELETE FROM notes WHERE id = ?;`, id); err != nil {
		return err
	}

	return nil
}
