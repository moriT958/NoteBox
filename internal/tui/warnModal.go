package tui

import (
	"charm.land/lipgloss/v2"
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
	confirm := m.styles.Modal.Confirm.Render(" (enter) Yes ")
	cancel := m.styles.Modal.Cancel.Render(" (ctrl+c) No ")
	tip := confirm + "           " + cancel
	modal := m.styles.Modal.Centered.
		Width(m.modalWidth).
		Height(m.modalHeight).
		Render("\n" + message + "\n\n" + tip)

	modal = m.styles.BorderActive.Render(modal)
	overlayX := m.width/2 - m.modalWidth/2
	overlayY := m.height/2 - m.modalHeight/2

	background := lipgloss.JoinVertical(lipgloss.Center,
		m.viewHeader(),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewListPanel(),
			m.viewPreviewer(),
		),
		m.viewHelp(),
	)

	fgLayer := lipgloss.NewLayer(modal).X(overlayX).Y(overlayY).Z(1)
	bgLayer := lipgloss.NewLayer(background).X(0).Y(0).Z(0)

	compositor := lipgloss.NewCompositor(bgLayer, fgLayer)
	canvas := lipgloss.NewCanvas(m.width, m.height).Compose(compositor)

	return m.styles.Main.Render(canvas.Render())
}
