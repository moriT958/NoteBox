package tui

import (
	"errors"
	"notebox/internal/note"
	"testing"

	"charm.land/bubbles/v2/viewport"
)

type fakeRenderer struct {
	rendered string
	err      error
	calls    int
}

func (f *fakeRenderer) RenderNote(note.Note) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.rendered, nil
}

func newTestPreviewer(width int, tabs ...*tab) *previewer {
	return &previewer{width: width, tabs: tabs}
}

func TestOpenTabCacheHit(t *testing.T) {
	existing := &tab{note: note.Note{Path: "a.md"}, rendered: "cached", isPreviewTab: true}
	p := newTestPreviewer(80, existing)
	r := &fakeRenderer{}
	p.renderer = r

	cmd := p.OpenTab(note.Note{Path: "a.md"}, false)

	if cmd != nil {
		t.Fatalf("expected no render command on cache hit, got one")
	}
	if r.calls != 0 {
		t.Fatalf("expected renderer not to be called on cache hit, got %d calls", r.calls)
	}
	if p.activeTab != 0 {
		t.Fatalf("activeTab = %d, want 0", p.activeTab)
	}
}

func TestOpenTabPromotesPreviewTabToPinned(t *testing.T) {
	existing := &tab{note: note.Note{Path: "a.md"}, rendered: "cached", isPreviewTab: true}
	p := newTestPreviewer(80, existing)

	if cmd := p.OpenTab(note.Note{Path: "a.md"}, true); cmd != nil {
		t.Fatalf("expected no render command on cache hit, got one")
	}
	if p.tabs[0].isPreviewTab {
		t.Fatalf("expected cached preview tab to be promoted to pinned")
	}
}

func TestOpenTabCacheMissRenders(t *testing.T) {
	p := newTestPreviewer(80)
	r := &fakeRenderer{rendered: "hello"}
	p.renderer = r

	cmd := p.OpenTab(note.Note{Path: "a.md"}, true)
	if cmd == nil {
		t.Fatalf("expected a render command on cache miss")
	}

	msg, ok := cmd().(tabRenderedMsg)
	if !ok {
		t.Fatalf("expected tabRenderedMsg, got %T", cmd())
	}
	if msg.note.Path != "a.md" || msg.rendered != "hello" || !msg.pin {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestOpenTabCacheMissRenderError(t *testing.T) {
	p := newTestPreviewer(80)
	p.renderer = &fakeRenderer{err: errors.New("boom")}

	cmd := p.OpenTab(note.Note{Path: "a.md"}, false)
	if _, ok := cmd().(errMsg); !ok {
		t.Fatalf("expected errMsg on render failure")
	}
}

func TestApplyRenderedPreviewReplacesExistingPreviewTab(t *testing.T) {
	old := &tab{note: note.Note{Path: "a.md"}, rendered: "old", isPreviewTab: true}
	p := newTestPreviewer(80, old)
	p.vp = viewport.New()

	p.applyRendered(tabRenderedMsg{note: note.Note{Path: "b.md"}, rendered: "new", pin: false})

	if len(p.tabs) != 1 {
		t.Fatalf("expected preview tab to be replaced in place, got %d tabs", len(p.tabs))
	}
	if p.tabs[0].note.Path != "b.md" || !p.tabs[0].isPreviewTab {
		t.Fatalf("unexpected tab state: %+v", p.tabs[0])
	}
}

func TestApplyRenderedPinnedAppendsWhenNoPreviewTabExists(t *testing.T) {
	pinned := &tab{note: note.Note{Path: "a.md"}, rendered: "a", isPreviewTab: false}
	p := newTestPreviewer(80, pinned)
	p.vp = viewport.New()

	p.applyRendered(tabRenderedMsg{note: note.Note{Path: "b.md"}, rendered: "b", pin: true})

	if len(p.tabs) != 2 {
		t.Fatalf("expected a new pinned tab to be appended, got %d tabs", len(p.tabs))
	}
	if p.activeTab != 1 {
		t.Fatalf("activeTab = %d, want 1", p.activeTab)
	}
}

func TestAdjustOffsetKeepsActiveTabVisible(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		tabCount   int
		activeTab  int
		offset     int
		wantOffset int
	}{
		{
			name:       "all tabs fit, no scroll needed",
			width:      100,
			tabCount:   3,
			activeTab:  2,
			offset:     0,
			wantOffset: 0,
		},
		{
			name:       "active tab beyond the visible window scrolls right",
			width:      maxTabWidth * 2,
			tabCount:   5,
			activeTab:  4,
			offset:     0,
			wantOffset: 3,
		},
		{
			name:       "active tab before the offset scrolls left",
			width:      maxTabWidth * 2,
			tabCount:   5,
			activeTab:  0,
			offset:     3,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabs := make([]*tab, tt.tabCount)
			for i := range tabs {
				tabs[i] = &tab{note: note.Note{Path: string(rune('a' + i))}}
			}
			p := newTestPreviewer(tt.width, tabs...)
			p.activeTab = tt.activeTab
			p.offset = tt.offset

			p.adjustOffset()

			if p.offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", p.offset, tt.wantOffset)
			}
			if p.activeTab < p.offset {
				t.Errorf("activeTab %d fell before offset %d", p.activeTab, p.offset)
			}
		})
	}
}
