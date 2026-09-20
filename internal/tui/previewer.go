package tui

import (
	"notebox/internal/tui/previewer"
	"notebox/internal/tui/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	borderH, _ := m.styles.BorderPassive.GetFrameSize()

	sidePanelWidth := msg.Width / layoutListPanelRatio
	contentWidth := msg.Width - sidePanelWidth - borderH*2
	contentHeight := msg.Height - helpGuideHeight - headerHeight - previewer.TabBarHeight - 1 // 1 is the connector height

	m.previewer.Width = contentWidth
	m.previewer.Height = max(1, contentHeight)

	m.previewer.VP.SetWidth(contentWidth)
	m.previewer.VP.SetHeight(max(1, contentHeight))
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

	tabBar := m.previewer.RenderTabBar(tabStyles, frameStyle)
	viewPort := border.UnsetBorderTop().Render(
		m.styles.Sized(m.previewer.Width, m.previewer.Height).Render(m.previewer.VP.View()),
	)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, viewPort)
}
