package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *model) handleListPanel(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return tea.Quit

	case "e":
		note := m.getNoteFromCursor()
		cmd = m.openFileWithEditor(note.GetFilePath())

	case "n":
		m.mode = createMode
		m.input.Focus()
		// cmd = textinput.Blink

	case "d":
		m.deleteNoteWarn()

	case "down", "j":
		m.listPanel.cursor++
		if m.listPanel.cursor >= len(m.listPanel.notes) {
			m.listPanel.cursor = 0
		}
		m.previewerRender()

	case "up", "k":
		m.listPanel.cursor--
		if m.listPanel.cursor < 0 {
			m.listPanel.cursor = len(m.listPanel.notes) - 1
		}
		m.previewerRender()

	case "right", "ctrl+l":
		m.focus = focusPreviewArea
	}

	return cmd
}
