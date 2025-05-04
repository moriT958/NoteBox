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

	notes, err := loadNoteFiles("./testdata")
	if err != nil {
		t.Fatal(err)
	}

	if len(notes) == 0 {
		t.Fatal("failed to load notes")
	}

	if len(notes) != 3 {
		t.Errorf("want 3 notes, but got %d", len(notes))
	}

	for i := range want {
		if want[i] != notes[i] {
			t.Errorf("want %v, but got %v\n", want[i], notes[i])
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
