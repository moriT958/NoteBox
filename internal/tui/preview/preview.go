package preview

import (
	"slices"
	"strings"

	"notebox/internal/core/note"
	"notebox/internal/tui/common"
	"notebox/internal/tui/styles"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	// maxTabWidth and minTabWidth include the tab's trailing separator.
	maxTabWidth = 20
	minTabWidth = 6
)

type tab struct {
	note     note.Note
	rendered string
	preview  bool
}

type Preview struct {
	com *common.Common

	// width and height include the border.
	width, height int

	tabs   []*tab
	active int
	// offset is the index of the first visible tab.
	offset int

	vp viewport.Model
}

func New(com *common.Common) *Preview {
	vp := viewport.New()
	vp.SetHorizontalStep(4)
	return &Preview{com: com, vp: vp}
}

func (p *Preview) SetSize(width, height int) {
	p.width, p.height = width, height
	// The tab bar replaces the top border, so only the sides and the bottom
	// border take space from the viewport.
	p.vp.SetWidth(max(0, width-2))
	p.vp.SetHeight(max(0, height-2))
	p.adjustOffset()
}

// ViewportKeyMap returns the scrolling keys, for help text.
func (p *Preview) ViewportKeyMap() viewport.KeyMap {
	return p.vp.KeyMap
}

// Open activates the tab showing n, pinning it if pin is set. It reports
// false when n has no tab yet, in which case the caller should render n and
// hand the result to SetRendered.
func (p *Preview) Open(n note.Note, pin bool) bool {
	i := p.index(n)
	if i < 0 {
		return false
	}
	if pin {
		p.tabs[i].preview = false
	}
	p.activate(i)
	return true
}

// SetRendered installs rendered content for n. An existing tab keeps its
// place and pin state; otherwise n becomes a new pinned tab or replaces the
// preview tab.
func (p *Preview) SetRendered(n note.Note, rendered string, pin bool) {
	if i := p.index(n); i >= 0 {
		p.tabs[i].rendered = rendered
		if pin {
			p.tabs[i].preview = false
		}
		if i == p.active {
			p.showActive(false)
		}
		return
	}

	t := &tab{note: n, rendered: rendered, preview: !pin}
	if i := slices.IndexFunc(p.tabs, func(t *tab) bool { return t.preview }); i >= 0 {
		p.tabs[i] = t
		p.activate(i)
		return
	}
	p.tabs = append(p.tabs, t)
	p.activate(len(p.tabs) - 1)
}

func (p *Preview) Has(n note.Note) bool {
	return p.index(n) >= 0
}

func (p *Preview) Active() (note.Note, bool) {
	if len(p.tabs) == 0 {
		return note.Note{}, false
	}
	return p.tabs[p.active].note, true
}

// Notes returns the notes of all open tabs.
func (p *Preview) Notes() []note.Note {
	notes := make([]note.Note, len(p.tabs))
	for i, t := range p.tabs {
		notes[i] = t.note
	}
	return notes
}

func (p *Preview) PinActive() {
	if len(p.tabs) > 0 {
		p.tabs[p.active].preview = false
	}
}

// CloseActive closes the active tab, always keeping at least one open.
func (p *Preview) CloseActive() {
	if len(p.tabs) <= 1 {
		return
	}
	p.removeAt(p.active)
}

// Remove closes the tab showing n, if any.
func (p *Preview) Remove(n note.Note) {
	if i := p.index(n); i >= 0 {
		p.removeAt(i)
	}
}

// Rename points the tab showing old at n.
func (p *Preview) Rename(old, n note.Note) {
	if i := p.index(old); i >= 0 {
		p.tabs[i].note = n
	}
}

func (p *Preview) Clear() {
	p.tabs = nil
	p.active = 0
	p.offset = 0
	p.vp.SetContent("")
}

func (p *Preview) NextTab() {
	if len(p.tabs) > 0 {
		p.activate((p.active + 1) % len(p.tabs))
	}
}

func (p *Preview) PrevTab() {
	if len(p.tabs) > 0 {
		p.activate((p.active - 1 + len(p.tabs)) % len(p.tabs))
	}
}

// Scroll forwards a key press to the viewport.
func (p *Preview) Scroll(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	p.vp, cmd = p.vp.Update(msg)
	return cmd
}

func (p *Preview) index(n note.Note) int {
	return slices.IndexFunc(p.tabs, func(t *tab) bool { return t.note == n })
}

func (p *Preview) activate(i int) {
	p.active = i
	p.adjustOffset()
	p.showActive(true)
}

func (p *Preview) removeAt(i int) {
	p.tabs = slices.Delete(p.tabs, i, i+1)
	if p.active > i || p.active >= len(p.tabs) {
		p.active = max(0, p.active-1)
	}
	if len(p.tabs) == 0 {
		p.Clear()
		return
	}
	p.adjustOffset()
	p.showActive(true)
}

func (p *Preview) showActive(top bool) {
	p.vp.SetContent(p.tabs[p.active].rendered)
	if top {
		p.vp.GotoTop()
	}
}

// tabWidths returns the widths of the tabs that fit in the tab bar when it
// starts at offset. Tabs are maxTabWidth wide; the last one may be cut down
// to what is left, as long as that is at least minTabWidth.
func (p *Preview) tabWidths(offset int) []int {
	// The bar starts with the left border "┃", and every tab ends with a
	// separator that is counted in its width.
	var widths []int
	used := 1
	for range p.tabs[offset:] {
		rem := p.width - used
		if rem < minTabWidth {
			break
		}
		w := min(maxTabWidth, rem)
		widths = append(widths, w)
		used += w
	}
	return widths
}

// adjustOffset scrolls the tab bar so the active tab is visible, and back
// left when tabs to the left would fit again (e.g. after closing a tab).
func (p *Preview) adjustOffset() {
	if len(p.tabs) == 0 || p.width == 0 {
		p.offset = 0
		return
	}
	p.offset = min(p.offset, p.active)
	for p.offset < p.active && p.active >= p.offset+len(p.tabWidths(p.offset)) {
		p.offset++
	}
	for p.offset > 0 && p.offset-1+len(p.tabWidths(p.offset-1)) >= len(p.tabs) {
		p.offset--
	}
}

func (p *Preview) Render(focused bool) string {
	sty := p.com.Styles
	frame, tabStyles, border := sty.FrameBlurred, sty.Tab.Blurred, sty.BorderBlurred
	if focused {
		frame, tabStyles, border = sty.FrameFocused, sty.Tab.Focused, sty.BorderFocused
	}

	body := border.UnsetBorderTop().
		Width(p.width).
		Height(p.height - 1).
		Render(p.vp.View())
	return lipgloss.JoinVertical(lipgloss.Left, p.renderTabBar(frame, tabStyles), body)
}

// renderTabBar draws the tabs as the pane's top border, e.g.
// "┃ Astro   ┃ Sample  ┣━━━━┓".
func (p *Preview) renderTabBar(frame lipgloss.Style, ts styles.TabStyles) string {
	if p.width < 2 {
		return ""
	}

	var widths []int
	if len(p.tabs) > 0 {
		widths = p.tabWidths(p.offset)
	}
	if len(widths) == 0 {
		return frame.Render("┏" + strings.Repeat("━", p.width-2) + "┓")
	}

	var b strings.Builder
	b.WriteString(frame.Render("┃"))
	used := 1
	for i, w := range widths {
		idx := p.offset + i
		t := p.tabs[idx]

		var s lipgloss.Style
		switch {
		case idx == p.active && t.preview:
			s = ts.ActivePreview
		case idx == p.active:
			s = ts.Active
		case t.preview:
			s = ts.InactivePreview
		default:
			s = ts.Inactive
		}

		// One column is the trailing separator; the rest is the label.
		inner := w - 1
		label := " " + ansi.Truncate(t.note.Title(), inner-1, "..")
		label += strings.Repeat(" ", max(0, inner-ansi.StringWidth(label)))
		b.WriteString(s.Render(label))
		used += w

		switch {
		case i < len(widths)-1:
			b.WriteString(frame.Render("┃"))
		case used == p.width:
			b.WriteString(frame.Render("┓"))
		default:
			b.WriteString(frame.Render("┣"))
		}
	}
	if rest := p.width - used; rest > 0 {
		b.WriteString(frame.Render(strings.Repeat("━", rest-1) + "┓"))
	}
	return b.String()
}
