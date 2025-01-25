package box

import (
	"os"
	"path/filepath"
	"testing"
)

var testDir = "testdata"

func TestSave_正常系(t *testing.T) {
	title := "test-2"
	wantFilename := "test-2.md"
	wantHeader := "# test-2"

	box := NewNoteBox(testDir)

	t.Run("successfully created new note", func(t *testing.T) {
		if err := box.Save(Note{Title: title}); err != nil {
			t.Fatalf("failed to create new file: %v", err)
		}

		notePath := filepath.Join(testDir, wantFilename)

		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			t.Errorf("file %s doesn't created", wantFilename)
		}

		content, err := os.ReadFile(notePath)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", notePath, err)
		}

		gotHeader := string(content)
		if gotHeader != wantHeader {
			t.Errorf("expected header: %s, got: %s", wantHeader, gotHeader)
		}

		defer func() {
			if err := os.Remove(notePath); err != nil {
				t.Fatal(err)
			}
		}()
	})
}

func TestSave_異常系(t *testing.T) {
	cases := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{"failed to create note, title already used", "test-1", true},
		{"failed to create note, empty title", "", true},
	}

	box := NewNoteBox(testDir)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := box.Save(Note{Title: c.title})
			if c.wantErr && err == nil {
				t.Error("expected failed to create note, but successed")
			}
		})
	}
}
