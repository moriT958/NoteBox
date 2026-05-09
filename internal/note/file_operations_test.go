package note

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestCreateNote(t *testing.T) {
	today := time.Now().Format(time.DateOnly)

	tests := []struct {
		name        string
		title       string
		wantTitle   string
		wantContent string
	}{
		{
			name:        "simple title",
			title:       "hello",
			wantTitle:   "hello",
			wantContent: "# hello\n\n",
		},
		{
			name:        "hyphenated title",
			title:       "my-note",
			wantTitle:   "my-note",
			wantContent: "# my-note\n\n",
		},
		{
			name:        "multi-word hyphenated title",
			title:       "hello-world-note",
			wantTitle:   "hello-world-note",
			wantContent: "# hello-world-note\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			got, err := CreateNote(dir, tt.title)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Title != tt.wantTitle {
				t.Errorf("Title: want %q, but got %q", tt.wantTitle, got.Title)
			}

			wantPath := filepath.Join(dir, fmt.Sprintf("%s-%s.md", tt.title, today))
			if got.Path != wantPath {
				t.Errorf("Path: want %q, but got %q", wantPath, got.Path)
			}

			if _, err := os.Stat(got.Path); os.IsNotExist(err) {
				t.Errorf("file does not exist: %s", got.Path)
			}

			content, err := os.ReadFile(got.Path)
			if err != nil {
				t.Fatalf("failed to read created file: %v", err)
			}
			if string(content) != tt.wantContent {
				t.Errorf("Content: want %q, but got %q", tt.wantContent, string(content))
			}
		})
	}
}

func TestRenameNote(t *testing.T) {
	today := time.Now().Format(time.DateOnly)

	tests := []struct {
		name      string
		note      Note
		newTitle  string
		wantTitle string
		wantPath  string
	}{
		{
			name:      "standard date suffix: title only changes",
			note:      Note{Title: "hello", Path: "/notes/hello-2025-05-02.md"},
			newTitle:  "world",
			wantTitle: "world",
			wantPath:  "/notes/world-2025-05-02.md",
		},
		{
			name:      "hyphenated original title: date suffix preserved",
			note:      Note{Title: "hi-there", Path: "/notes/hi-there-2025-05-04.md"},
			newTitle:  "new-title",
			wantTitle: "new-title",
			wantPath:  "/notes/new-title-2025-05-04.md",
		},
		{
			name:      "new title with hyphens: date suffix preserved",
			note:      Note{Title: "hello", Path: "/notes/hello-2025-05-02.md"},
			newTitle:  "my-new-note",
			wantTitle: "my-new-note",
			wantPath:  "/notes/my-new-note-2025-05-02.md",
		},
		{
			name:      "no date suffix: today appended",
			note:      Note{Title: "todo", Path: "/notes/todo.md"},
			newTitle:  "tasks",
			wantTitle: "tasks",
			wantPath:  "/notes/tasks-" + today + ".md",
		},
		{
			name:      "date-only filename (3 segments): today appended",
			note:      Note{Title: "2025-05-02", Path: "/notes/2025-05-02.md"},
			newTitle:  "diary",
			wantTitle: "diary",
			wantPath:  "/notes/diary-" + today + ".md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renameNote(tt.note, tt.newTitle)
			if got.Title != tt.wantTitle {
				t.Errorf("Title: want %q, but got %q", tt.wantTitle, got.Title)
			}
			if got.Path != tt.wantPath {
				t.Errorf("Path: want %q, but got %q", tt.wantPath, got.Path)
			}
		})
	}
}
