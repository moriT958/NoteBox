package tui

import (
	"notebox/models"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) handleWarnModal(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.String() {
	case "y":
		cmd = m.deleteNote()
		m.previewerRender()

	case "n", "esc":
		m.warnModal.open = false
		m.warnModal.message = ""
	}

	return cmd
}

func (m *model) deleteNote() tea.Cmd {
	var cmd tea.Cmd

	if err := os.Remove(m.listPanel.notes[m.listPanel.cursor].GetFilePath()); err != nil {
		cmd = func() tea.Msg { return errMsg{err} }
	}

	id := m.listPanel.notes[m.listPanel.cursor].ID
	models.GetRepository().DeleteByID(id)
	m.listPanel.notes = slices.Delete(m.listPanel.notes, m.listPanel.cursor, m.listPanel.cursor+1)
	if m.listPanel.cursor > 0 && m.listPanel.cursor >= len(m.listPanel.notes) {
		m.listPanel.cursor--
	}
	m.warnModal.open = false
	m.warnModal.message = ""

	return cmd
}
