package note

import (
	"errors"
	"os"
	"testing"
)

func TestNoteService_WriteNoteContent(t *testing.T) {
	svc := NewNoteService(newTestBox(t))
	n, err := svc.CreateNote("memo")
	if err != nil {
		t.Fatal(err)
	}
	_, tag, err := svc.ReadNote(n)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("writes when the ETag matches", func(t *testing.T) {
		newTag, err := svc.WriteNoteContent(n, []byte("# memo\n\nv2\n"), tag)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if newTag == tag {
			t.Errorf("ETag did not change")
		}
		got, _ := os.ReadFile(svc.AbsPath(n))
		if string(got) != "# memo\n\nv2\n" {
			t.Errorf("content = %q", got)
		}
		tag = newTag
	})

	t.Run("refuses a stale ETag and reports the current one", func(t *testing.T) {
		cur, err := svc.WriteNoteContent(n, []byte("lost"), "stale")
		if !errors.Is(err, ErrConflict) {
			t.Fatalf("err = %v, want ErrConflict", err)
		}
		if cur != tag {
			t.Errorf("current ETag = %q, want %q", cur, tag)
		}
	})

	t.Run("writes unconditionally without an ETag", func(t *testing.T) {
		if _, err := svc.WriteNoteContent(n, []byte("forced"), ""); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}
