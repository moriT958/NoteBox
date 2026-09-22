package fuzzymodal

import (
	"notebox/internal/notedeprecated"
	"testing"
)

func TestFilterNotesEmptyQueryReturnsAllItems(t *testing.T) {
	items := []notedeprecated.Note{{Title: "a"}, {Title: "b"}}

	got := filterNotes("", items)

	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
}

func TestFilterNotesMatchesByTitle(t *testing.T) {
	items := []notedeprecated.Note{{Title: "grocery-list"}, {Title: "todo"}, {Title: "groceries-2"}}

	got := filterNotes("groc", items)

	if len(got) != 2 {
		t.Fatalf("got %d items, want 2: %+v", len(got), got)
	}
	for _, n := range got {
		if n.Title != "grocery-list" && n.Title != "groceries-2" {
			t.Errorf("unexpected match: %+v", n)
		}
	}
}

func TestFilterResetsCursorAndOffset(t *testing.T) {
	m := &Modal{AllItems: []notedeprecated.Note{{Title: "a"}, {Title: "b"}}, Cursor: 5, Offset: 2}

	m.Filter("")

	if m.Cursor != 0 || m.Offset != 0 {
		t.Fatalf("Filter did not reset cursor/offset: cursor=%d offset=%d", m.Cursor, m.Offset)
	}
}
