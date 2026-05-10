package tui

import (
	"notebox/internal/config"
	"notebox/internal/note"

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
	}, nil
}

const (
	// this all includes border size
	// tabBarHeight = border_top(1) + border_bottom(0) + inner(1)
	tabBarHeight = 2
	// maxTabWidth = border_side(2) + inner(18)
	maxTabWidth = 20
	// minTabWidth = border_side(2) + inner(4)
	minTabWidth = 6
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	borderH, borderV := m.styles.BorderPassive.GetFrameSize()

	sidePanelWidth := msg.Width / layoutListPanelRatio
	contentWidth := msg.Width - sidePanelWidth - borderH*2
	contentHeight := msg.Height - borderV - helpGuideHeight - tabBarHeight

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
	// If already exist	in tabs cache, activate it.
	for i, t := range p.tabs {
		if !t.isPreviewTab && t.note.Path == note.Path {
			p.activeTab = i
			p.vp.SetContent(p.tabs[i].rendered)
			p.vp.GotoTop()
			return nil
		}
	}
	// If not exsist in tabs cache, then fire rendering job.
	return renderPreviewCmd(p.renderer, note)
}

// openTab opens selected note on previewer.
func (p *previewer) openTab(msg openNormalTabMsg) {
	msg.isPreviewTab = false
	p.tabs[p.activeTab] = (*tab)(&msg)
}

/*
 * TODO: ↓ All WIP
 */

func (m model) viewPreviewer() string {
	var (
		tabBar   string
		viewPort string
	)

	if m.focus == onPreviewer {
		tabBar = m.previewer.renderTabBar(m.styles.TabBarFocused)
		viewPort = m.styles.BorderActive.UnsetBorderTop().Render(
			m.styles.Sized(m.previewer.width, m.previewer.height).Render(m.previewer.vp.View()),
		)
	} else {
		tabBar = m.previewer.renderTabBar(m.styles.TabBarUnforcused)
		viewPort = m.styles.BorderPassive.UnsetBorderTop().Render(
			m.styles.Sized(m.previewer.width, m.previewer.height).Render(m.previewer.vp.View()),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, viewPort)
}

func (p previewer) calcLastTabWidth() int {
	tabCount := len(p.tabs)
	tabBarWidth := tabCount * maxTabWidth
	if tabBarWidth > p.width {
		truncateWidth := tabBarWidth - p.width
		return maxTabWidth - truncateWidth
	}
	return maxTabWidth
}

func (p previewer) isActive(t tab) bool { return t == *p.tabs[p.activeTab] }

func (p previewer) renderTabBar(style styles.TabBarStyles) string {
}

func truncateTabLabel(label string, max int) string {
	width := 0
	result := []rune{}

	for _, r := range label {
		rw := runewidth.RuneWidth(r)
		if width+rw > max {
			return string(result) + ".."
		}
		width += rw
		result = append(result, r)
	}

	return label
}
