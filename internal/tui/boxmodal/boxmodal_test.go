package boxmodal

import (
	"notebox/internal/note"
	"testing"
)

func TestApplyBoxesLoadedFocusesCurrentBox(t *testing.T) {
	m := &BoxModal{}
	boxes := []note.Box{{ID: 1}, {ID: 2}, {ID: 3}}

	m.ApplyBoxesLoaded(boxes, 2)

	if len(m.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(m.Items))
	}
	if m.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (index of box ID 2)", m.Cursor)
	}
}

func TestApplyBoxCreatedAppends(t *testing.T) {
	m := &BoxModal{Items: []note.Box{{ID: 1}}}

	m.ApplyBoxCreated(note.Box{ID: 2, Title: "new"})

	if len(m.Items) != 2 || m.Items[1].Title != "new" {
		t.Fatalf("unexpected items: %+v", m.Items)
	}
}

func TestApplyBoxDeletedKeepsCursorInBounds(t *testing.T) {
	m := &BoxModal{Items: []note.Box{{ID: 1}, {ID: 2}, {ID: 3}}, Cursor: 2}

	m.ApplyBoxDeleted(3)

	if len(m.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(m.Items))
	}
	if m.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after deleting the last item", m.Cursor)
	}
}

func TestApplyBoxRenamedReplacesInPlace(t *testing.T) {
	m := &BoxModal{Items: []note.Box{{ID: 1, Title: "old"}, {ID: 2, Title: "b"}}}

	m.ApplyBoxRenamed(note.Box{ID: 1, Title: "renamed"})

	if m.Items[0].Title != "renamed" {
		t.Fatalf("items[0].Title = %q, want %q", m.Items[0].Title, "renamed")
	}
	if m.Items[1].Title != "b" {
		t.Fatalf("unexpected mutation of unrelated item: %+v", m.Items[1])
	}
}
