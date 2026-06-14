package tui

import (
	tea "charm.land/bubbletea/v2"
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
	h, _ := m.styles.Main.GetFrameSize()
	m.input.Placeholder = "Enter note name..."
	m.input.Focus()
	m.input.CharLimit = 50
	m.input.SetWidth((msg.Width - h) / 3)
}

func (m model) viewTypingModal() string {
	confirm := m.styles.Modal.Confirm.Render(" (" + selectionModalConfirmKey + ") Create ")
	cancel := m.styles.Modal.Cancel.Render(" (" + selectionModalCancelKey + ") Cancel ")
	tip := confirm + "           " + cancel
	modal := m.styles.Modal.Centered.
		Width(m.modalWidth).
		Height(m.modalHeight).
		Render("\n" + m.input.View() + "\n\n" + tip)

	modalX := (m.width - m.modalWidth) / 2
	modalY := (m.height - m.modalHeight) / 2
	modal = m.styles.BorderActive.Render(modal)

	return m.renderOverlay(modal, modalX, modalY)
}
