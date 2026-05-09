package note

import (
	"testing"
)

func TestRenderNote_EmptyNote(t *testing.T) {
	renderer, err := NewGlamourRenderer("dark")
	if err != nil {
		t.Fatalf("NewGlamourRenderer: %v", err)
	}

	got, err := renderer.RenderNote(Note{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "(( No Content ))"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
