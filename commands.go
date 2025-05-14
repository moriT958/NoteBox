package main

import tea "github.com/charmbracelet/bubbletea"

/* Messages */

type errMsg struct{ err error }

type renderPreviewMsg struct{ path string }
type createNewNoteMsg struct{ note note }
type typingModalMsg struct{ isOpen bool }

/* Commands */

func errCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}
}

func toggleModalCmd(state bool) tea.Cmd {
	return func() tea.Msg {
		if state {
			return typingModalMsg{true}
		} else {
			return typingModalMsg{false}
		}
	}
}
