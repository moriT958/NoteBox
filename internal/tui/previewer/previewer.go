// Package previewer owns the note preview pane: its tabs, the single
// render-and-cache path behind them, and the tab bar's layout arithmetic.
package previewer

import (
	"notebox/internal/config"
	"notebox/internal/note"
	"notebox/internal/tui/styles"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

// TabBarHeight is a single row that doubles as the viewport's top border.
const TabBarHeight = 1

const (
	// maxTabWidth = separator(1) + inner(19)
	maxTabWidth = 20
	// minTabWidth = separator(1) + inner(5)
	minTabWidth = 6
)

type Previewer struct {
	Width, Height int
	VP            viewport.Model
	Renderer      note.NoteRenderer
	// tabs include normal tabs and preview tabs.
	tabs []*tab
	// currently active tab index
	activeTab int
	// tab scroll offset: index of the first visible tab
	offset int
	keys   KeyMap
}

// KeyMap is the previewer's own key bindings. DefaultKeyMap is also used by
// the root tui package to build help text, so its fields are exported.
type KeyMap struct {
	FocusList    key.Binding
	EditNote     key.Binding
	Up           key.Binding
	Down         key.Binding
	OpenTab      key.Binding
	CloseTab     key.Binding
	NextTab      key.Binding
	PrevTab      key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
}

func DefaultKeyMap() KeyMap {
	vpKeys := viewport.DefaultKeyMap()
	return KeyMap{
		FocusList: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "list"),
		),
		EditNote: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Up:   vpKeys.Up,
		Down: vpKeys.Down,
		OpenTab: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open tab"),
		),
		CloseTab: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "close tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		HalfPageUp:   vpKeys.HalfPageUp,
		HalfPageDown: vpKeys.HalfPageDown,
	}
}

// FocusListRequestedMsg reports that the user asked to move focus back to
// the note list while the previewer was focused.
type FocusListRequestedMsg struct{}

// EditRequestedMsg reports that the user asked to open the active tab's
// note in an external editor. The root owns the editor command since it
// holds the configured editor.
type EditRequestedMsg struct {
	Path string
}

type tab struct {
	note     note.Note
	rendered string
	// preview tab is an unpinned tab that has not been fully opened yet.
	isPreviewTab bool
}

func New(cfg *config.Config) (*Previewer, error) {
	vp := viewport.New(
		viewport.WithWidth(0),
		viewport.WithHeight(0),
	)
	vp.SetHorizontalStep(4)

	r, err := note.NewGlamourRenderer(cfg.Theme)
	if err != nil {
		return nil, err
	}

	return &Previewer{
		VP:        vp,
		Renderer:  r,
		tabs:      []*tab{},
		activeTab: 0,
		offset:    0,
		keys:      DefaultKeyMap(),
	}, nil
}

// SetSize resizes the previewer and its viewport.
func (p *Previewer) SetSize(width, height int) {
	p.Width = width
	p.Height = height
	p.VP.SetWidth(width)
	p.VP.SetHeight(height)
}

// activeNote returns the note behind the currently active tab, or the zero
// value if there are no tabs.
func (p *Previewer) activeNote() note.Note {
	if len(p.tabs) == 0 {
		return note.Note{}
	}
	return p.tabs[p.activeTab].note
}

// PinActiveTab promotes the currently active tab from an ephemeral preview
// tab to a pinned normal tab, if it isn't already.
func (p *Previewer) PinActiveTab() {
	if len(p.tabs) == 0 {
		return
	}
	p.tabs[p.activeTab].isPreviewTab = false
}

// Update handles a key press while the previewer is focused.
func (p *Previewer) Update(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, p.keys.FocusList):
		return func() tea.Msg { return FocusListRequestedMsg{} }
	case key.Matches(msg, p.keys.EditNote):
		path := p.activeNote().Path
		return func() tea.Msg { return EditRequestedMsg{Path: path} }
	case key.Matches(msg, p.keys.OpenTab):
		p.PinActiveTab()
		return nil
	case key.Matches(msg, p.keys.CloseTab):
		p.CloseTab()
		return nil
	case key.Matches(msg, p.keys.NextTab):
		p.NextTab()
		return nil
	case key.Matches(msg, p.keys.PrevTab):
		p.PrevTab()
		return nil
	default:
		var cmd tea.Cmd
		p.VP, cmd = p.VP.Update(msg)
		return cmd
	}
}

// errMsg reports a render failure.
type errMsg error

// TabRenderedMsg carries the rendered content for a note the previewer asked
// to be rendered. Pin reports whether the result should become a pinned
// (normal) tab or stay an ephemeral preview tab.
type TabRenderedMsg struct {
	Note     note.Note
	Rendered string
	Pin      bool
}

// renderTabCmd is the previewer's single render entry point: every path that
// needs a note rendered into a tab (cursor movement, opening a tab, resize,
// note creation/rename/deletion) goes through this one command.
func renderTabCmd(renderer note.NoteRenderer, n note.Note, pin bool) tea.Cmd {
	return func() tea.Msg {
		rendered, err := renderer.RenderNote(n)
		if err != nil {
			return errMsg(err)
		}
		return TabRenderedMsg{Note: n, Rendered: rendered, Pin: pin}
	}
}

// Refresh forces a fresh render of n into an ephemeral preview tab,
// bypassing the tab cache. Used when the rendered content itself may be
// stale (e.g. a resize, or the note's file just changed on disk).
func (p *Previewer) Refresh(n note.Note) tea.Cmd {
	return renderTabCmd(p.Renderer, n, false)
}

// OpenTab is the previewer's single entry point for showing a note: cursor
// movement (pin=false, ephemeral preview) and the "open tab" key (pin=true,
// pinned normal tab) both go through here. An already-cached tab for n is
// activated in place (promoted to pinned if requested); otherwise a render
// job is fired and ApplyRendered installs the result on completion.
func (p *Previewer) OpenTab(n note.Note, pin bool) tea.Cmd {
	for i, t := range p.tabs {
		if t.note.Path == n.Path {
			if pin && t.isPreviewTab {
				t.isPreviewTab = false
			}
			p.activeTab = i
			p.adjustOffset()
			p.updateViewportContent()
			return nil
		}
	}
	return renderTabCmd(p.Renderer, n, pin)
}

// ApplyRendered installs the result of a renderTabCmd job as either an
// ephemeral preview tab or a pinned normal tab, and refreshes the viewport.
func (p *Previewer) ApplyRendered(msg TabRenderedMsg) {
	newTab := &tab{note: msg.Note, rendered: msg.Rendered, isPreviewTab: !msg.Pin}
	if newTab.isPreviewTab {
		p.setPreviewTab(newTab)
	} else {
		p.setNormalTab(newTab)
	}
	p.updateViewportContent()
}

func (p *Previewer) setPreviewTab(prevTab *tab) {
	// validate prevTab is preview tab.
	if !prevTab.isPreviewTab {
		return
	}

	for i, t := range p.tabs {
		if t.isPreviewTab {
			p.tabs[i] = prevTab
			p.activeTab = i
			p.adjustOffset()
			return
		}
	}

	p.tabs = append(p.tabs, prevTab)
	p.activeTab = len(p.tabs) - 1
	p.adjustOffset()
}

// setNormalTab installs newTab as a pinned tab, replacing an existing
// preview-tab slot if one exists, or appending otherwise.
func (p *Previewer) setNormalTab(newTab *tab) {
	for i, t := range p.tabs {
		if t.isPreviewTab {
			p.tabs[i] = newTab
			p.activeTab = i
			p.adjustOffset()
			return
		}
	}
	p.tabs = append(p.tabs, newTab)
	p.activeTab = len(p.tabs) - 1
	p.adjustOffset()
}

// updateViewportContent updates the viewport content to match the active tab.
func (p *Previewer) updateViewportContent() {
	if len(p.tabs) == 0 {
		return
	}
	p.VP.SetContent(p.tabs[p.activeTab].rendered)
	p.VP.GotoTop()
}

// CloseTab removes the currently active tab and updates activeTab/offset.
// At least one tab is always kept regardless of preview/normal status.
func (p *Previewer) CloseTab() {
	if len(p.tabs) <= 1 {
		return
	}
	p.tabs = slices.Delete(p.tabs, p.activeTab, p.activeTab+1)
	if p.activeTab >= len(p.tabs) {
		p.activeTab = len(p.tabs) - 1
	}
	p.adjustOffset()
	p.updateViewportContent()
}

// ClearAllTabs resets the tab list and viewport content.
func (p *Previewer) ClearAllTabs() {
	p.tabs = []*tab{}
	p.activeTab = 0
	p.offset = 0
	p.VP.SetContent("")
}

// RemoveTabByPath removes a tab whose note matches the given path.
func (p *Previewer) RemoveTabByPath(path string) {
	for i, t := range p.tabs {
		if t.note.Path == path {
			p.tabs = slices.Delete(p.tabs, i, i+1)
			if p.activeTab >= len(p.tabs) && p.activeTab > 0 {
				p.activeTab--
			}
			p.adjustOffset()
			p.updateViewportContent()
			return
		}
	}
}

// NextTab moves the active tab one step to the right (wraps around).
func (p *Previewer) NextTab() {
	if len(p.tabs) == 0 {
		return
	}
	p.activeTab = (p.activeTab + 1) % len(p.tabs)
	p.adjustOffset()
	p.updateViewportContent()
}

// PrevTab moves the active tab one step to the left (wraps around).
func (p *Previewer) PrevTab() {
	if len(p.tabs) == 0 {
		return
	}
	p.activeTab = (p.activeTab - 1 + len(p.tabs)) % len(p.tabs)
	p.adjustOffset()
	p.updateViewportContent()
}

// TODO: Refactor this.
// adjustOffset ensures the active tab is within the visible range,
// and reduces offset when tabs fit into a smaller window (e.g. after deletion).
func (p *Previewer) adjustOffset() {
	n := len(p.tabs)
	if n == 0 || p.Width == 0 {
		p.offset = 0
		return
	}

	if p.activeTab < p.offset {
		p.offset = p.activeTab
	}

	// Find the last visible tab index from current offset.
	usedW := 0
	lastVisible := p.offset
	for i := p.offset; i < n; i++ {
		rem := p.Width - usedW
		if rem < minTabWidth {
			break
		}
		usedW += min(maxTabWidth, rem)
		lastVisible = i
	}

	if p.activeTab > lastVisible {
		// Active tab is beyond visible range: find offset from the right.
		usedW = 0
		p.offset = p.activeTab
		for j := p.activeTab; j >= 0; j-- {
			usedW += maxTabWidth
			if usedW > p.Width {
				p.offset = j + 1
				break
			}
			p.offset = j
		}
		return
	}

	// Active tab is visible. Reclaim space on the left: find the minimum offset
	// that still keeps the last tab visible (handles tab deletion shrinking the list).
	if lastVisible == n-1 && p.offset > 0 {
		usedW = 0
		newOffset := n - 1
		for j := n - 1; j >= 0; j-- {
			if usedW+maxTabWidth > p.Width {
				break
			}
			usedW += maxTabWidth
			newOffset = j
		}
		if newOffset < p.offset {
			p.offset = newOffset
		}
	}
}

// View renders the previewer's tab bar and viewport, styled according to
// whether the previewer currently has focus.
func (p *Previewer) View(s *styles.Style, focused bool) string {
	var (
		tabStyles  styles.TabBarStyles
		frameStyle lipgloss.Style
		border     lipgloss.Style
	)
	if focused {
		tabStyles = s.TabBarFocused
		frameStyle = s.ActiveColor
		border = s.BorderActive
	} else {
		tabStyles = s.TabBarUnforcused
		frameStyle = s.PassiveColor
		border = s.BorderPassive
	}

	tabBar := p.RenderTabBar(tabStyles, frameStyle)
	viewPort := border.UnsetBorderTop().Render(
		s.Sized(p.Width, p.Height).Render(p.VP.View()),
	)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, viewPort)
}

// TODO: Refactor this function.
// RenderTabBar draws a single-row tab bar that doubles as the viewport's top
// border, e.g. "┃ Astro ┃ Sample ┣━━━┓". Tabs share a single "┃" separator
// (no double borders between them); the row closes with "┣"+fill+"┓" if tabs
// don't reach the right edge, or a bare "┓" if they do.
func (p *Previewer) RenderTabBar(tabStyles styles.TabBarStyles, frameStyle lipgloss.Style) string {
	// Total visual width includes the viewport's left/right borders, because
	// this row spans the full frame width.
	W := p.Width + 2
	if W < 2 {
		return ""
	}

	emptyBar := func() string {
		return frameStyle.Render("┏" + strings.Repeat("━", W-2) + "┓")
	}

	if len(p.tabs) == 0 {
		return emptyBar()
	}

	type layout struct {
		start, width    int
		active, preview bool
		label           string
	}

	// Pass 1: figure out which tabs are visible and where they sit.
	var visible []layout
	pos := 0
	for i := p.offset; i < len(p.tabs); i++ {
		t := p.tabs[i]
		rem := W - pos
		if rem < minTabWidth {
			break
		}
		tabW := min(maxTabWidth, rem)
		visible = append(visible, layout{
			start:   pos,
			width:   tabW,
			active:  i == p.activeTab,
			preview: t.isPreviewTab,
			label:   t.note.Title,
		})
		pos += tabW
	}

	if len(visible) == 0 {
		return emptyBar()
	}

	var bar strings.Builder
	bar.WriteString(frameStyle.Render("┃"))
	cursor := 1

	for vi, b := range visible {
		// Fill gap (rendered with frame color since it stands in for the frame).
		if b.start+1 > cursor {
			gapW := b.start + 1 - cursor
			bar.WriteString(frameStyle.Render(strings.Repeat("━", gapW)))
			cursor += gapW
		}

		var ts lipgloss.Style
		switch {
		case b.active && b.preview:
			ts = tabStyles.ActivePreview
		case b.active:
			ts = tabStyles.Active
		case b.preview:
			ts = tabStyles.InactivePreview
		default:
			ts = tabStyles.Inactive
		}

		// One column of b.width is reserved for the trailing separator/corner
		// written below; the rest is the label.
		innerW := b.width - 1
		labelText := " " + truncateTabLabel(b.label, innerW-1)
		if lw := runewidth.StringWidth(labelText); lw < innerW {
			labelText += strings.Repeat(" ", innerW-lw)
		}
		bar.WriteString(ts.Render(labelText))
		cursor += innerW

		isLast := vi == len(visible)-1
		atRight := cursor+1 == W
		switch {
		case isLast && atRight:
			bar.WriteString(frameStyle.Render("┓"))
		case isLast:
			bar.WriteString(frameStyle.Render("┣"))
		default:
			bar.WriteString(frameStyle.Render("┃"))
		}
		cursor++
	}

	// Fill remainder to the right of the last tab, closing the viewport's
	// top-right corner with ┓.
	if cursor < W {
		remW := W - cursor
		if remW == 1 {
			bar.WriteString(frameStyle.Render("┓"))
		} else {
			bar.WriteString(frameStyle.Render(strings.Repeat("━", remW-1) + "┓"))
		}
	}

	return bar.String()
}

// truncateTabLabel returns a string whose visual width does not exceed max tab width.
func truncateTabLabel(label string, max int) string {
	if max <= 0 {
		return ""
	}

	if runewidth.StringWidth(label) <= max {
		return label
	}

	const ellipsis = ".."
	const ellipsisWidth = 2

	// not enough space to append ".."
	if max < ellipsisWidth {
		return truncateString(label, max)
	}

	return truncateString(label, max-ellipsisWidth) + ellipsis
}

// truncateString truncates to the specified width
func truncateString(str string, limit int) string {
	width := 0
	result := []rune{}
	for _, r := range str {
		rw := runewidth.RuneWidth(r)
		if width+rw > limit {
			break
		}
		width += rw
		result = append(result, r)
	}
	return string(result)
}
