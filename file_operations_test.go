package main

import (
	"strings"
	"testing"
)

func TestLoadNoteFiles(t *testing.T) {
	want := []note{
		{"hello", "# hello\n\n"},
		{"nice", "# nice\n\n"},
		{"test1", "# test1\n\n"},
	}

	cmd := loadNoteFiles("./testdata")
	msg := cmd()
	if m, ok := msg.(errMsg); ok {
		t.Fatal(m.err)
	}

	if m, ok := msg.(notesLoadedMsg); ok {
		if len(m.notes) == 0 {
			t.Fatal("failed to load notes")
		}

		if len(m.notes) != 3 {
			t.Errorf("want 3 notes, but got %d", len(m.notes))
		}

		for i := range want {
			if want[i] != m.notes[i] {
				t.Errorf("want %v, but got %v\n", want[i], m.notes[i])
			}
		}
	}
}

func TestGetTitleFromFilename(t *testing.T) {
	files := []string{
		"hello-2025-05-02.md",
		"nice-2025-05-02.md",
		"test1-2025-05-02.md",
	}

	want := []string{"hello", "nice", "test1"}

	for i := range files {
		got := getTitleFromFilename(files[i])
		if strings.Compare(got, want[i]) != 0 {
			t.Errorf("want %s, but got %s\n", want[i], got)
		}
	}
}
