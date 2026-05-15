package tui

import (
	"notebox/internal/config"
	"notebox/internal/note"
	"notebox/internal/tui/styles"
	"slices"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
)

type previewer struct {
	width, height int
	vp            viewport.Model
	renderer      note.NoteRenderer
	// tabs include normal tabs and preview tabs.
	tabs []*tab
	// currently active tab index
	activeTab int
	// tab scroll offset: index of the first visible tab
	offset int
}

type tab struct {
	note     note.Note
	rendered string
	// preview tab is an unpinned tab that has not been fully opened yet.
	isPreviewTab bool
}

func newPreviewer(cfg *config.Config) (*previewer, error) {
	vp := viewport.New(
		viewport.WithWidth(0),
		viewport.WithHeight(0),
	)
	vp.SetHorizontalStep(4)

	r, err := note.NewGlamourRenderer(cfg.Theme)
	if err != nil {
		return nil, err
	}

	return &previewer{
		vp:        vp,
		renderer:  r,
		tabs:      []*tab{},
		activeTab: 0,
		offset:    0,
	}, nil
}

const (
	// tabBarHeight = top_border(1) + label(1) + connector_border(1)
	// the connector row also serves as the viewport's top border
	tabBarHeight = 3
	// maxTabWidth = border_side(2) + inner(18)
	maxTabWidth = 20
	// minTabWidth = border_side(2) + inner(4)
	minTabWidth = 6
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	borderH, _ := m.styles.BorderPassive.GetFrameSize()

	sidePanelWidth := msg.Width / layoutListPanelRatio
	contentWidth := msg.Width - sidePanelWidth - borderH*2
	contentHeight := msg.Height - helpGuideHeight - tabBarHeight - 1 // 1 is the connector height

	m.previewer.width = contentWidth
	m.previewer.height = max(1, contentHeight)

	m.previewer.vp.SetWidth(contentWidth)
	m.previewer.vp.SetHeight(max(1, contentHeight))
}

func (m *model) updatePreviewerContent(msg renderPreviewMsg) {
	// Update preview tab
	newPreviewTab := &tab{
		note:         m.listPanel.selectedItem(),
		rendered:     string(msg),
		isPreviewTab: true,
	}
	m.previewer.setPreviewTab(newPreviewTab)

	// Update viewport content
	m.previewer.vp.SetContent(string(msg))
	m.previewer.vp.GotoTop()
}

func (p *previewer) setPreviewTab(prevTab *tab) {
	// validate prevTab is preview tab.
	if !prevTab.isPreviewTab {
		return
	}

	for i, t := range p.tabs {
		if t.isPreviewTab {
			p.tabs[i] = prevTab
			p.activeTab = i
			return
		}
	}

	p.tabs = append(p.tabs, prevTab)
	p.activeTab = len(p.tabs) - 1
}

// previewNote sets rendered note content on previewer.
func (p *previewer) previewNote(note note.Note) tea.Cmd {
	// If already exist in tabs cache, activate it.
	for i, t := range p.tabs {
		if !t.isPreviewTab && t.note.Path == note.Path {
			p.activeTab = i
			p.vp.SetContent(p.tabs[i].rendered)
			p.vp.GotoTop()
			return nil
		}
	}
	// If not exist in tabs cache, then fire rendering job.
	return renderPreviewCmd(p.renderer, note)
}

// openTab promotes the current preview tab (or adds) to a normal tab.
func (p *previewer) openTab(msg openNormalTabMsg) {
	msg.isPreviewTab = false
	p.tabs[p.activeTab] = (*tab)(&msg)
}

// updateViewportContent updates the viewport content to match the active tab.
func (p *previewer) updateViewportContent() {
	if len(p.tabs) == 0 {
		return
	}
	p.vp.SetContent(p.tabs[p.activeTab].rendered)
	p.vp.GotoTop()
}

// closeTab removes the currently active tab and updates activeTab/offset.
// At least one tab is always kept regardless of preview/normal status.
func (p *previewer) closeTab() {
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

// removeTabByPath removes a tab whose note matches the given path.
func (p *previewer) removeTabByPath(path string) {
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

// nextTab moves the active tab one step to the right (wraps around).
func (p *previewer) nextTab() {
	if len(p.tabs) == 0 {
		return
	}
	p.activeTab = (p.activeTab + 1) % len(p.tabs)
	p.adjustOffset()
	p.updateViewportContent()
}

// prevTab moves the active tab one step to the left (wraps around).
func (p *previewer) prevTab() {
	if len(p.tabs) == 0 {
		return
	}
	p.activeTab = (p.activeTab - 1 + len(p.tabs)) % len(p.tabs)
	p.adjustOffset()
	p.updateViewportContent()
}

// adjustOffset ensures the active tab is within the visible range.
func (p *previewer) adjustOffset() {
	n := len(p.tabs)
	if n == 0 || p.width == 0 {
		p.offset = 0
		return
	}

	if p.activeTab < p.offset {
		p.offset = p.activeTab
		return
	}

	// Find the last visible tab index from current offset.
	usedW := 0
	lastVisible := p.offset
	for i := p.offset; i < n; i++ {
		rem := p.width - usedW
		if rem < minTabWidth {
			break
		}
		usedW += min(maxTabWidth, rem)
		lastVisible = i
	}

	if p.activeTab <= lastVisible {
		return
	}

	// Active tab is beyond visible range: find offset from the right.
	usedW = 0
	p.offset = p.activeTab
	for j := p.activeTab; j >= 0; j-- {
		usedW += maxTabWidth
		if usedW > p.width {
			p.offset = j + 1
			break
		}
		p.offset = j
	}
}

func (m model) viewPreviewer() string {
	var (
		tabStyles  styles.TabBarStyles
		frameStyle lipgloss.Style
		border     lipgloss.Style
	)
	if m.focus == onPreviewer {
		tabStyles = m.styles.TabBarFocused
		frameStyle = m.styles.ActiveColor
		border = m.styles.BorderActive
	} else {
		tabStyles = m.styles.TabBarUnforcused
		frameStyle = m.styles.PassiveColor
		border = m.styles.BorderPassive
	}

	tabBar := m.previewer.renderTabBar(tabStyles, frameStyle)
	viewPort := border.UnsetBorderTop().Render(
		m.styles.Sized(m.previewer.width, m.previewer.height).Render(m.previewer.vp.View()),
	)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, viewPort)
}

// TODO: Refactor this function.
// renderTabBar draws a 3-row tab bar.
// The bottom row (connector) doubles as the viewport's top edge:
// it uses ┻/┣/┫ where non-active tab edges meet the frame, and ┛/┗ around the active tab so its bottom opens into the viewport.
func (p previewer) renderTabBar(tabStyles styles.TabBarStyles, frameStyle lipgloss.Style) string {
	// Total visual width includes the viewport's left/right borders, because
	// the connector row spans the full frame width.
	W := p.width + 2
	if W < 2 {
		return ""
	}

	emptyBar := func() string {
		blank := strings.Repeat(" ", W)
		top := frameStyle.Render("┏" + strings.Repeat("━", W-2) + "┓")
		return strings.Join([]string{blank, blank, top}, "\n")
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
	// Tabs are flush so adjacent borders read as ┓┏ / ┃┃ / ┻┻ (or ┗┻ etc.
	// around the active tab).
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

	var top, lab, conn strings.Builder
	cursor := 0

	for vi, b := range visible {
		// Fill gap (rendered with frame color since the connector is frame).
		if b.start > cursor {
			gapW := b.start - cursor
			top.WriteString(strings.Repeat(" ", gapW))
			lab.WriteString(strings.Repeat(" ", gapW))
			conn.WriteString(frameStyle.Render(strings.Repeat("━", gapW)))
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

		// Row 0: ┏━...━┓
		top.WriteString(ts.Render("┏" + strings.Repeat("━", b.width-2) + "┓"))

		// Row 1: ┃ <label padded> ┃
		innerW := b.width - 2
		labelText := " " + truncateTabLabel(b.label, innerW-1)
		if lw := runewidth.StringWidth(labelText); lw < innerW {
			labelText += strings.Repeat(" ", innerW-lw)
		}
		lab.WriteString(ts.Render("┃" + labelText + "┃"))

		// Row 2: connector — corners depend on adjacency to outer frame.
		isFirst := vi == 0
		isLast := vi == len(visible)-1
		atLeft := b.start == 0
		atRight := b.start+b.width == W

		var left, right, inside string
		if b.active {
			inside = strings.Repeat(" ", b.width-2)
			if isFirst && atLeft {
				// active tab shares the outer left frame
				left = "┃"
			} else {
				left = "┛"
			}
			if isLast && atRight {
				right = "┃"
			} else {
				right = "┗"
			}
		} else {
			inside = strings.Repeat("━", b.width-2)
			if isFirst && atLeft {
				left = "┣"
			} else {
				left = "┻"
			}
			if isLast && atRight {
				right = "┫"
			} else {
				right = "┻"
			}
		}
		// The connector row is logically the viewport's top edge, so colour it
		// with the frame style regardless of the tab's own colour.
		conn.WriteString(frameStyle.Render(left + inside + right))

		cursor = b.start + b.width
	}

	// Fill remainder to the right of the last tab. The very last column closes
	// the viewport's top-right corner with ┓.
	if cursor < W {
		remW := W - cursor
		top.WriteString(strings.Repeat(" ", remW))
		lab.WriteString(strings.Repeat(" ", remW))
		if remW == 1 {
			conn.WriteString(frameStyle.Render("┓"))
		} else {
			conn.WriteString(frameStyle.Render(strings.Repeat("━", remW-1) + "┓"))
		}
	}

	return strings.Join([]string{top.String(), lab.String(), conn.String()}, "\n")
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
