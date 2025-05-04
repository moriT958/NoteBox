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
	switch msg := cmd().(type) {
	case errMsg:
		t.Fatal(msg.err)
	case notesLoadedMsg:
		if len(msg.notes) == 0 {
			t.Fatal("failed to load notes")
		}

		if len(msg.notes) != 3 {
			t.Errorf("want 3 notes, but got %d", len(msg.notes))
		}

		for i := range want {
			if want[i] != msg.notes[i] {
				t.Errorf("want %v, but got %v\n", want[i], msg.notes[i])
			}
		}
	}
}

func TestGetTitleFromFilename(t *testing.T) {
	files := []string{
		"hello-2025-05-02.md",
		"nice-2025-05-02.md",
		"test1-2025-05-02.md",
		"hi-there-2025-05-04.md",
	}

	want := []string{"hello", "nice", "test1", "hi-there"}

	for i := range files {
		got := getTitleFromFilename(files[i])
		if strings.Compare(got, want[i]) != 0 {
			t.Errorf("want %s, but got %s\n", want[i], got)
		}
	}
}
