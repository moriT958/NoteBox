package typingmodal

import "github.com/charmbracelet/lipgloss"

var (
	ModalConfirm = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#99d1db"))
	ModalCancel  = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#ea999c"))
)
