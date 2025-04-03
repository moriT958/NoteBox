package models

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestSave(t *testing.T) {
	db, err := sql.Open("sqlite", "../data.sqlite")
	if err != nil {
		t.Fatal(err)
	}

	repo, err := NewNoteRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	note := Note{
		ID:        1,
		Title:     "Test",
		CreatedAt: time.Now(),
	}

	id, err := repo.Save(note)
	if err != nil {
		t.Error(err)
	}
	if id != note.ID {
		t.Error("idが違います")
	}
}
