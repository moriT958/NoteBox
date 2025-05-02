package main

import "testing"

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
