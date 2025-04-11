package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *model) handlePreviewer(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.String() {
	case "left", "ctrl+h":
		m.focus = focusListArea

	case "ctrl+c", "q", "esc":
		return tea.Quit

	default:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return cmd
}
