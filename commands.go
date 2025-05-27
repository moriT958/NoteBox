package main

import tea "github.com/charmbracelet/bubbletea"

/* Messages */

type errMsg struct{ err error }

type renderPreviewMsg struct{ path string }

/* Commands */

func errCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errMsg{err}
	}
}
