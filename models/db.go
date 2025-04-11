package models

import (
	"database/sql"
	"time"
)

const initQuery = `
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL UNIQUE,
	created_at TEXT NOT NULL
);`

// NoteRepository implements Repository interface
// to check that, comment out below.
// var _ Repository = (*NoteRepository)(nil)
type NoteRepository struct {
	DB *sql.DB
}

func NewNoteRepository(db *sql.DB) error {
	// initial database schema apply
	if _, err := db.Exec(initQuery); err != nil {
		return err
	}
	repository := &NoteRepository{DB: db}
	nr = repository
	return nil
}

func (r *NoteRepository) Save(note Note) (int, error) {
	if err := r.DB.QueryRow(`INSERT INTO notes (title, created_at) VALUES (?, ?) RETURNING id;`,
		note.Title, note.CreatedAtStr()).Scan(&note.ID); err != nil {
		return 0, err
	}
	return note.ID, nil
}

func (r *NoteRepository) FindByID(id int) (*Note, error) {
	note := new(Note)

	var timeStr string
	if err := r.DB.QueryRow(`SELECT id, title, created_at FROM notes WHERE id = ?;`, id).
		Scan(&note.ID, &note.Title, &timeStr); err != nil {
		return &Note{}, err
	}
	t, err := time.Parse("2006-01-02", timeStr)
	if err != nil {
		return nil, err
	}
	note.CreateAt = t

	return note, nil
}

func (r *NoteRepository) FindAll() ([]*Note, error) {
	notes := make([]*Note, 0)
	rows, err := r.DB.Query(`SELECT * FROM notes;`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		n := new(Note)
		var timeStr string
		if err := rows.Scan(&n.ID, &n.Title, &timeStr); err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02", timeStr)
		if err != nil {
			return nil, err
		}
		n.CreateAt = t
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *NoteRepository) DeleteByID(id int) error {
	if _, err := r.DB.Exec(`DELETE FROM notes WHERE id = ?;`, id); err != nil {
		return err
	}
	return nil
}
