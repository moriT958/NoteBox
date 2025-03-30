package models

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"testing"
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
		ID:    1,
		Title: "Test",
		Path:  "./test/path",
	}

	id, err := repo.Save(note)
	if err != nil {
		t.Error(err)
	}
	if id != note.ID {
		t.Error("idが違います")
	}
}
