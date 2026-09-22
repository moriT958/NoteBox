package tui

import (
	"notebox/internal/notedeprecated"
	"notebox/internal/tui/listpanel"
	"notebox/internal/tui/previewer"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
)

type fakeRenderer struct {
	rendered string
}

func (f *fakeRenderer) RenderNote(notedeprecated.Note) (string, error) {
	return f.rendered, nil
}

// TestHandleWarnModalKeysDeleteNoteOrder guards the invariant documented at
// the deleteNoteFileCmd call site: the selected path must be captured before
// removeItem runs, since removeItem mutates the list cursor that
// SelectedItem() depends on. Driving the handler directly (no tea.Program,
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
		listPanel: listpanel.ListPanel{
			Items:  []notedeprecated.Note{{Title: "keep", Path: keep}, {Title: "gone", Path: gone}},
			Cursor: 1,
		},
		previewer: previewer.Previewer{Renderer: &fakeRenderer{rendered: "ok"}},
	}

	cmd := m.handleWarnModalKeys(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a batched command")
	}

	if len(m.listPanel.Items) != 1 || m.listPanel.Items[0].Path != keep {
		t.Fatalf("expected only %q to remain, got %+v", keep, m.listPanel.Items)
	}
	if m.listPanel.Cursor != 0 {
		t.Fatalf("cursor = %d, want 0 after removing the last item", m.listPanel.Cursor)
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
		listPanel: listpanel.ListPanel{
			Items:  []notedeprecated.Note{{Title: "a", Path: "a.md"}, {Title: "b", Path: "b.md"}},
			Cursor: 0,
			Height: 10,
		},
		previewer: previewer.Previewer{Renderer: &fakeRenderer{rendered: "ok"}},
	}

	cmd := m.handleListPanelKeys(tea.KeyPressMsg{Code: 'j', Text: "j"})

	if m.listPanel.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.listPanel.Cursor)
	}
	if cmd == nil {
		t.Fatal("expected OpenTab to return a render command for the newly selected item")
	}
	msg, ok := cmd().(previewer.TabRenderedMsg)
	if !ok {
		t.Fatalf("expected previewer.TabRenderedMsg, got %T", cmd())
	}
	if msg.Note.Path != "b.md" || msg.Pin {
		t.Fatalf("unexpected message: %+v", msg)
	}
}
