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

	got, err := loadNoteFiles("./testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("failed to load notes")
	}

	if noteNum := len(got); noteNum != 3 {
		t.Errorf("want 3 notes, but got %d", noteNum)
	}

	for i := range want {
		if want[i] != got[i] {
			t.Errorf("want %v, but got %v\n", want[i], got[i])
		}
	}
}

// func TestWalk(t *testing.T) {
// 	wantFileNum := 3
// 	files := make([]string, 0)
// 	if err := filepath.Walk("./testdata", func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}
//
// 		_, filename := filepath.Split(path)
// 		files = append(files, filename)
//
// 		return nil
// 	}); err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if len(files) != wantFileNum {
// 		t.Errorf("want %d, but got %d\n", wantFileNum, len(files))
// 	}
// }

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
