package tui

import (
	"errors"
	"notebox/models"

	tea "github.com/charmbracelet/bubbletea"
)

func StartApp() error {
	// Get notes from repository
	notes, err := models.GetRepository().FindAll()
	if err != nil {
		return err
	}

	m := initModel(notes)
	if m == nil {
		return errors.New("fail to initialize bubbletea model")
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
