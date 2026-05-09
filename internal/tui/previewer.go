package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	sidePanelWidth := msg.Width / layoutSidePanelRatio
	previewWidth := msg.Width - sidePanelWidth - layoutFramePadding
	contentHeight := msg.Height - layoutFramePadding - helpGuideHeight

	m.vp.SetWidth(previewWidth)
	m.vp.SetHeight(max(1, contentHeight))
}

func (m *model) updatePreviewerContent(msg renderPreviewMsg) tea.Cmd {
	m.vp.SetContent(string(msg))
	m.vp.GotoTop()
	return nil
}

func (m model) viewPreviewer() string {
	view := strings.Builder{}
	view.WriteString(m.vp.View())

	if m.focus == onPreviewer {
		return m.styles.BorderActive.Render(
			m.styles.Sized(m.vp.Width(), m.vp.Height()).Render(view.String()))
	}
	return m.styles.BorderPassive.Render(
		m.styles.Sized(m.vp.Width(), m.vp.Height()).Render(view.String()))
}

