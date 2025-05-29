package tui

import (
	stringfunction "notebox/internal/pkg/string_function"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) toggleWarnModal(ac modalAction) {
	switch ac {
	case open:
		m.focus = onWarnModal
	case shut:
		m.focus = onListPanel
	}
}

func (m model) viewWarnModal() string {

	message := "Are you sure you want to remove?"
	confirm := m.styles.modalConfirmColor.Render(" (enter) Yes ")
	cancel := m.styles.modalCalcelColor.Render(" (ctrl+c) No ")
	tip := confirm + "           " + cancel
	modal := lipgloss.NewStyle().Width(m.modalWidth).Height(m.modalHeight).
		Align(lipgloss.Center, lipgloss.Center).Render("\n" + message + "\n\n" + tip)

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
