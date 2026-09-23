package note

import (
	"os"
	"path/filepath"
	"testing"

	"notebox/internal/core/box"
)

func newTestBox(t *testing.T) box.Box {
	t.Helper()
	return box.Box{ID: "test", Title: "Test Box", Path: t.TempDir(), Active: true}
}

func TestNoteService_GetNotes(t *testing.T) {
	t.Run("Successfully get only markdown notes.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "note1.md"), []byte("# note1"), 0644)
		os.WriteFile(filepath.Join(box.Path, "note2.md"), []byte("# note2"), 0644)
		os.WriteFile(filepath.Join(box.Path, "ignore.txt"), []byte("not a note"), 0644)

		svc := NewNoteService(box)

		got, err := svc.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("len(got) = %d, want 2", len(got))
		}
	})

	t.Run("Returns empty slice, when box has no notes.", func(t *testing.T) {
		box := newTestBox(t)
		svc := NewNoteService(box)

		got, err := svc.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(got) = %d, want 0", len(got))
		}
	})
}

func TestNoteService_CreateNote(t *testing.T) {
	t.Run("Successfully create note file.", func(t *testing.T) {
		box := newTestBox(t)
		s := NewNoteService(box)

		_, err := s.CreateNote("New Note")
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(box.Path, "New Note.md"))
		if err != nil {
			t.Fatalf("expected file to be created: %v", err)
		}
		if string(content) != "# New Note\n\n" {
			t.Errorf("content = %q, want %q", string(content), "# New Note")
		}
	})

	t.Run("Fail to create note, when a note with the same title already exists.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "New Note.md"), []byte("# original content"), 0644)

		s := NewNoteService(box)

		_, err := s.CreateNote("New Note")
		if err == nil {
			t.Fatalf("error was expected, but not occured.")
		}

		content, err := os.ReadFile(filepath.Join(box.Path, "New Note.md"))
		if err != nil {
			t.Fatalf("expected existing file to still exist: %v", err)
		}
		if string(content) != "# original content" {
			t.Errorf("existing note content was overwritten: got %q", string(content))
		}
	})
}

func TestNoteService_CreateNote_TitleRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{name: "japanese", title: "買い物リスト"},
		{name: "slash", title: "Q1/Q2 Planning"},
		{name: "colon", title: "TODO: buy milk"},
		{name: "backslash and colon", title: `C:\Users\test`},
		{name: "percent sign", title: "100% done"},
		{name: "question mark", title: "is this ok?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := newTestBox(t)
			s := NewNoteService(box)

			if _, err := s.CreateNote(tt.title); err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}

			notes, err := s.GetNotes()
			if err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}
			if len(notes) != 1 {
				t.Fatalf("expected 1 note, got %d", len(notes))
			}
			if notes[0].title != tt.title {
				t.Errorf("title = %q, want %q", notes[0].title, tt.title)
			}
		})
	}
}

func TestNoteService_GetNoteContent(t *testing.T) {
	t.Run("Successfully get note file content.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "note1.md"), []byte("# note1 content"), 0644)

		s := NewNoteService(box)

		notes, err := s.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}

		got, err := s.GetNoteContent(notes[0])
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if got != "# note1 content" {
			t.Errorf("content = %q, want %q", got, "# note1 content")
		}
	})
}

func TestNoteService_RemoveNote(t *testing.T) {
	t.Run("Successfull remove note file.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "note1.md"), []byte("# note1"), 0644)

		s := NewNoteService(box)

		notes, err := s.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}

		if err := s.RemoveNote(notes[0]); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		if _, err := os.Stat(filepath.Join(box.Path, "note1.md")); !os.IsNotExist(err) {
			t.Errorf("expected note1.md to be removed, stat err = %v", err)
		}
	})
}

func TestNoteService_RenameNote(t *testing.T) {
	t.Run("Successfully rename note file name.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "note1.md"), []byte("# note1"), 0644)

		s := NewNoteService(box)

		notes, err := s.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}

		if _, err := s.RenameNote(notes[0], "renamed"); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		if _, err := os.Stat(filepath.Join(box.Path, "renamed.md")); err != nil {
			t.Errorf("expected renamed.md to exist: %v", err)
		}
		if _, err := os.Stat(filepath.Join(box.Path, "note1.md")); !os.IsNotExist(err) {
			t.Errorf("expected note1.md to no longer exist, stat err = %v", err)
		}
	})

	t.Run("Successfully rename note file name in a subdirectory, keeping its directory.", func(t *testing.T) {
		box := newTestBox(t)
		subDir := filepath.Join(box.Path, "sub")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}
		os.WriteFile(filepath.Join(subDir, "note1.md"), []byte("# note1"), 0644)

		s := NewNoteService(box)

		notes, err := s.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}

		if _, err := s.RenameNote(notes[0], "renamed"); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		if _, err := os.Stat(filepath.Join(subDir, "renamed.md")); err != nil {
			t.Errorf("expected renamed.md to exist in subdir: %v", err)
		}
		if _, err := os.Stat(filepath.Join(subDir, "note1.md")); !os.IsNotExist(err) {
			t.Errorf("expected note1.md to no longer exist, stat err = %v", err)
		}
	})

	t.Run("Fail to rename note, when a note with the new title already exists.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "note1.md"), []byte("# note1"), 0644)
		os.WriteFile(filepath.Join(box.Path, "note2.md"), []byte("# note2 original"), 0644)

		s := NewNoteService(box)

		notes, err := s.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		var note1 Note
		for _, n := range notes {
			if n.title == "note1" {
				note1 = n
			}
		}

		if _, err := s.RenameNote(note1, "note2"); err == nil {
			t.Fatalf("error was expected, but not occured.")
		}

		if _, err := os.Stat(filepath.Join(box.Path, "note1.md")); err != nil {
			t.Errorf("expected note1.md to still exist: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(box.Path, "note2.md"))
		if err != nil {
			t.Fatalf("expected note2.md to still exist: %v", err)
		}
		if string(content) != "# note2 original" {
			t.Errorf("existing note2.md content was overwritten: got %q", string(content))
		}
	})

	t.Run("Successfully renames to the same title as a no-op.", func(t *testing.T) {
		box := newTestBox(t)
		os.WriteFile(filepath.Join(box.Path, "note1.md"), []byte("# note1"), 0644)

		s := NewNoteService(box)

		notes, err := s.GetNotes()
		if err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}
		if len(notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(notes))
		}

		if _, err := s.RenameNote(notes[0], "note1"); err != nil {
			t.Fatalf("unexpected err occurred: %v", err)
		}

		if _, err := os.Stat(filepath.Join(box.Path, "note1.md")); err != nil {
			t.Errorf("expected note1.md to still exist: %v", err)
		}
	})
}
