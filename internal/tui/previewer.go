package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	m.vp.Height = msg.Height * 5 / 6
	m.vp.Width = msg.Width * 2 / 3
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
		return m.styles.borderActive.Render(
			adjustSize(m.vp.Width, m.vp.Height)(view.String()))
	}
	return m.styles.borderPassive.Render(
		adjustSize(m.vp.Width, m.vp.Height)(view.String()))
}
