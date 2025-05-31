package tui

import (
	stringfunction "notebox/internal/pkg/string_function"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type modalAction int

const (
	open modalAction = iota
	shut
)

func (m *model) toggleTypingModal(ac modalAction) {
	switch ac {
	case open:
		m.input.Reset()
		m.focus = onTypingModal
	case shut:
		m.focus = onListPanel
	}
}

func (m *model) updateTypingModalSize(msg tea.WindowSizeMsg) {
	h, _ := m.styles.main.GetFrameSize()
	m.input.Placeholder = "Enter note name..."
	m.input.Focus()
	m.input.CharLimit = 50
	m.input.Width = (msg.Width - h) / 3
}

func (m model) viewTypingModal() string {
	confirm := m.styles.modalConfirmColor.Render(" (enter) Create ")
	cancel := m.styles.modalCalcelColor.Render(" (ctrl+c) Cancel ")
	tip := confirm + "           " + cancel
	modal := lipgloss.NewStyle().Width(m.modalWidth).Height(m.modalHeight).
		Align(lipgloss.Center, lipgloss.Center).Render("\n" + m.input.View() + "\n\n" + tip)

	modal = m.styles.borderActive.Render(modal)
	overlayX := m.width/2 - m.modalWidth/2
	overlayY := m.height/2 - m.modalHeight/2
	background := lipgloss.JoinVertical(lipgloss.Center,
		m.viewHeader(),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewListPanel(),
			m.viewPreviewer()))
	return m.styles.main.Render(
		stringfunction.PlaceOverlay(overlayX, overlayY, modal, background))
}
