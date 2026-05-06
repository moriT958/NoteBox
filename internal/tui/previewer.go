package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	m.vp.SetWidth(msg.Width - msg.Width/4 - 4)
	m.vp.SetHeight(msg.Height - 4)
}

func (m *model) updatePreviewerContent(msg renderPreviewMsg) tea.Cmd {
	content := string(msg)
	renderedContent, err := m.renderer.Render(string(content))
	if err != nil {
		return errCmd(err)
	}
	m.vp.SetContent(renderedContent)
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
