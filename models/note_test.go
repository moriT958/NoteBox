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

	if err := NewNoteRepository(db); err != nil {
		t.Fatal(err)
	}

	repo := GetRepository()
	if err != nil {
		t.Fatal(err)
	}

	note := Note{
		ID:       1,
		Title:    "Test",
		CreateAt: time.Now(),
	}

	id, err := repo.Save(note)
	if err != nil {
		t.Error(err)
	}
	if id != note.ID {
		t.Error("idが違います")
	}
}
