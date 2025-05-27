package typingmodal

import tea "github.com/charmbracelet/bubbletea"

type InputDataMsg string

// this is used to get textinput data.
func (m Model) inputCmd() tea.Cmd {
	return func() tea.Msg { return InputDataMsg(m.input.Value()) }
}

type TypingModalMsg bool

// this is used to toggle(open/close) modal.
func ToggleModalCmd(open bool) tea.Cmd {
	return func() tea.Msg { return TypingModalMsg(open) }
}
