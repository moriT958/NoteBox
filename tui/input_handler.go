package tui

import (
	"fmt"
	"notebox/models"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) handleInputModal(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		if m.mode == createMode {
			var note *models.Note
			title := m.input.Value()
			note, cmd = m.createNote(title)

			m.listPanel.notes = append(m.listPanel.notes, note)
			m.listPanel.cursor = len(m.listPanel.notes) - 1

			m.previewerRender()
		}
		m.input.SetValue("")
		m.input.Blur()

	case "esc", "ctrl+c":
		m.input.SetValue("")
		m.mode = navigateMode
		m.input.Blur()
	}
	m.input, cmd = m.input.Update(msg)

	return cmd
}

func (m *model) createNote(title string) (*models.Note, tea.Cmd) {
	note := &models.Note{
		Title:    title,
		CreateAt: time.Now(),
	}
	id, err := models.GetRepository().Save(*note)
	if err != nil {
		return nil, func() tea.Msg { return errMsg{err} }
	}

	note, err = models.GetRepository().FindByID(id)
	if err != nil {
		return nil, func() tea.Msg { return errMsg{err} }
	}

	topHeader := "# " + title + "\n\n"
	fp, err := os.Create(note.GetFilePath())
	if err != nil {
		return nil, func() tea.Msg { return errMsg{err} }
	}
	defer fp.Close()
	fmt.Fprint(fp, topHeader)

	return note, nil
}
