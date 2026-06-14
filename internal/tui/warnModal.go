package tui

func (m *model) toggleWarnModal(ac modalAction) {
	switch ac {
	case open:
		m.focus = onWarnModal
	case shut:
		m.focus = onListPanel
	}
}

func (m model) viewWarnModal() string {
	message := m.warnMessage
	confirm := m.styles.Modal.Confirm.Render(" (" + selectionModalConfirmKey + ") Yes ")
	cancel := m.styles.Modal.Cancel.Render(" (" + selectionModalCancelKey + ") No ")
	tip := confirm + "           " + cancel
	modal := m.styles.Modal.Centered.
		Width(m.modalWidth).
		Height(m.modalHeight).
		Render("\n" + message + "\n\n" + tip)

	modal = m.styles.BorderActive.Render(modal)
	overlayX := m.width/2 - m.modalWidth/2
	overlayY := m.height/2 - m.modalHeight/2

	return m.renderOverlay(modal, overlayX, overlayY)
}
