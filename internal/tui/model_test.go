package tui

import (
	"notebox/internal/note"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestHandleWarnModalKeysDeleteNoteOrder guards the invariant documented at
// the deleteNoteFileCmd call site: the selected path must be captured before
// removeItem runs, since removeItem mutates the list cursor that
// selectedItem() depends on. Driving the handler directly (no tea.Program,
// no full model) is exactly the win the per-focus split is for.
func TestHandleWarnModalKeysDeleteNoteOrder(t *testing.T) {
	dir := t.TempDir()
	keep := filepath.Join(dir, "keep.md")
	if err := os.WriteFile(keep, []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	gone := filepath.Join(dir, "gone.md")
	if err := os.WriteFile(gone, []byte("gone"), 0644); err != nil {
		t.Fatal(err)
	}

	m := &model{
		focus:      onWarnModal,
		warnAction: warnDeleteNote,
		keys:       defaultKeyMap(),
		listPanel: listPanel{
			items:  []note.Note{{Title: "keep", Path: keep}, {Title: "gone", Path: gone}},
			cursor: 1,
		},
		previewer: previewer{renderer: &fakeRenderer{rendered: "ok"}},
	}

	cmd := m.handleWarnModalKeys(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a batched command")
	}

	if len(m.listPanel.items) != 1 || m.listPanel.items[0].Path != keep {
		t.Fatalf("expected only %q to remain, got %+v", keep, m.listPanel.items)
	}
	if m.listPanel.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 after removing the last item", m.listPanel.cursor)
	}
	if m.focus != onListPanel {
		t.Fatalf("focus = %v, want onListPanel", m.focus)
	}

	// Execute the deferred batch to confirm deleteNoteFileCmd was built with
	// the path captured *before* removeItem ran, not the (now different)
	// currently-selected item.
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", cmd())
	}
	for _, c := range batch {
		c()
	}
	if _, err := os.Stat(gone); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be removed from disk", gone)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("expected %q to survive, got err: %v", keep, err)
	}
}

func TestHandleListPanelKeysCursorMovesPreview(t *testing.T) {
	m := &model{
		focus: onListPanel,
		keys:  defaultKeyMap(),
		listPanel: listPanel{
			items:  []note.Note{{Title: "a", Path: "a.md"}, {Title: "b", Path: "b.md"}},
			cursor: 0,
			height: 10,
		},
		previewer: previewer{renderer: &fakeRenderer{rendered: "ok"}},
	}

	cmd := m.handleListPanelKeys(tea.KeyPressMsg{Code: 'j', Text: "j"})

	if m.listPanel.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.listPanel.cursor)
	}
	if cmd == nil {
		t.Fatal("expected OpenTab to return a render command for the newly selected item")
	}
	msg, ok := cmd().(tabRenderedMsg)
	if !ok {
		t.Fatalf("expected tabRenderedMsg, got %T", cmd())
	}
	if msg.note.Path != "b.md" || msg.pin {
		t.Fatalf("unexpected message: %+v", msg)
	}
}
