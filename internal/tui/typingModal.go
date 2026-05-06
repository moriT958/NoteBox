package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	confirm := m.styles.Modal.Confirm.Render(" (enter) Create ")
	cancel := m.styles.Modal.Cancel.Render(" (ctrl+c) Cancel ")
	tip := confirm + "           " + cancel
	modal := m.styles.Modal.Centered.
		Width(m.modalWidth).
		Height(m.modalHeight).
		Render("\n" + m.input.View() + "\n\n" + tip)

	modalX := (m.width - m.modalWidth) / 2
	modalY := (m.height - m.modalHeight) / 2
	modal = m.styles.BorderActive.Render(modal)

	background := lipgloss.JoinVertical(lipgloss.Center,
		m.viewHeader(),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewListPanel(),
			m.viewPreviewer()))

	fgLayer := lipgloss.NewLayer(modal).X(modalX).Y(modalY).Z(1)
	bgLayer := lipgloss.NewLayer(background).X(0).Y(0).Z(0)

	compositor := lipgloss.NewCompositor(bgLayer, fgLayer)
	canvas := lipgloss.NewCanvas(m.width, m.height).Compose(compositor)

	return m.styles.Main.Render(canvas.Render())
}
