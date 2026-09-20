package tui

import (
	"notebox/internal/note"
	"testing"
)

func TestApplyBoxesLoadedFocusesCurrentBox(t *testing.T) {
	m := &boxModal{}
	boxes := []note.Box{{ID: 1}, {ID: 2}, {ID: 3}}

	m.applyBoxesLoaded(boxes, 2)

	if len(m.items) != 3 {
		t.Fatalf("items = %d, want 3", len(m.items))
	}
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (index of box ID 2)", m.cursor)
	}
}

func TestApplyBoxCreatedAppends(t *testing.T) {
	m := &boxModal{items: []note.Box{{ID: 1}}}

	m.applyBoxCreated(note.Box{ID: 2, Title: "new"})

	if len(m.items) != 2 || m.items[1].Title != "new" {
		t.Fatalf("unexpected items: %+v", m.items)
	}
}

func TestApplyBoxDeletedKeepsCursorInBounds(t *testing.T) {
	m := &boxModal{items: []note.Box{{ID: 1}, {ID: 2}, {ID: 3}}, cursor: 2}

	m.applyBoxDeleted(3)

	if len(m.items) != 2 {
		t.Fatalf("items = %d, want 2", len(m.items))
	}
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after deleting the last item", m.cursor)
	}
}

func TestApplyBoxRenamedReplacesInPlace(t *testing.T) {
	m := &boxModal{items: []note.Box{{ID: 1, Title: "old"}, {ID: 2, Title: "b"}}}

	m.applyBoxRenamed(note.Box{ID: 1, Title: "renamed"})

	if m.items[0].Title != "renamed" {
		t.Fatalf("items[0].Title = %q, want %q", m.items[0].Title, "renamed")
	}
	if m.items[1].Title != "b" {
		t.Fatalf("unexpected mutation of unrelated item: %+v", m.items[1])
	}
}
