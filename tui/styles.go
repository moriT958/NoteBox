package tui

import "github.com/charmbracelet/lipgloss"

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	docStyle  = lipgloss.NewStyle().Margin(0, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
)
