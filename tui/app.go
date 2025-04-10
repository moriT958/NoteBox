package tui

import tea "github.com/charmbracelet/bubbletea"

func StartApp() error {
	m, err := initModel()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
