package note

import (
	"strings"
	"testing"
)

func TestLoadNoteFiles(t *testing.T) {
	want := []Note{
		{"hello", "../testdata/notes/hello-2025-05-02.md"},
		{"nice", "../testdata/notes/nice-2025-05-02.md"},
		{"test1", "../testdata/notes/test1-2025-05-02.md"},
	}

	notes, err := LoadNoteFiles("../testdata/notes")
	if err != nil {
		t.Fatal(err)
	}

	if len(notes) == 0 {
		t.Fatal("failed to load notes")
	}

	if len(notes) != len(want) {
		t.Fatalf("want %d notes, but got %d", len(want), len(notes))
	}

	for i := range want {
		if want[i] != notes[i] {
			t.Errorf("want %v, but got %v\n", want[i], notes[i])
		}
	}
}

func TestGetTitleFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "standard date suffix", filename: "hello-2025-05-02.md", want: "hello"},
		{name: "another standard", filename: "nice-2025-05-02.md", want: "nice"},
		{name: "numeric title", filename: "test1-2025-05-02.md", want: "test1"},
		{name: "hyphenated title", filename: "hi-there-2025-05-04.md", want: "hi-there"},
		{name: "short filename without date", filename: "todo.md", want: "todo"},
		{name: "three segments only", filename: "2025-05-02.md", want: "2025-05-02"},
		{name: "no extension", filename: "memo", want: "memo"},
		{name: "four segments but no date format", filename: "a-b-c-d.md", want: "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTitleFromFilename(tt.filename)
			if strings.Compare(got, tt.want) != 0 {
				t.Errorf("want %s, but got %s", tt.want, got)
			}
		})
	}
}
