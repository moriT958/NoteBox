package models

import (
	"database/sql"
	"notebox/config"
	"path/filepath"
	"time"
)

type Note struct {
	ID        int
	Title     string
	CreatedAt time.Time
}

func (n *Note) CreatedAtStr() string {
	return n.CreatedAt.Format(`2006-01-02`)
}

func (n *Note) GetFilePath() string {
	return filepath.Join(config.Volume, n.Title+"-"+n.CreatedAtStr()+".md")
}

type Repository interface {
	Save(Note) (int, error)
	FindByID(int) (*Note, error)
	FindAll() ([]*Note, error)
	DeleteByID(int) error
}

type NoteRepository struct {
	DB *sql.DB
}

const initQuery = `
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL UNIQUE,
	created_at TEXT NOT NULL
);
`

func NewNoteRepository(db *sql.DB) (*NoteRepository, error) {
	// Migrate database
	if _, err := db.Exec(initQuery); err != nil {
		return nil, err
	}
	return &NoteRepository{DB: db}, nil
}

// Implement Repository interface
var _ Repository = (*NoteRepository)(nil)

// TODO:
// add update methods. if exsits, override.
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
	note.CreatedAt = t

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
		n.CreatedAt = t

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
