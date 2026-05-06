package tui

func (m model) viewHelp() string {
	if m.help.ShowAll {
		return ""
	}

	guideFocus := onListPanel
	if m.focus == onPreviewer {
		guideFocus = onPreviewer
	}

	focused := m.keys.forFocus(guideFocus)
	guide := m.help.ShortHelpView(focused.ShortHelp())
	return m.styles.Help.
		Width(m.width).
		Render(guide)
}

func (m model) viewFullHelpOverlay() string {
	if !m.help.ShowAll {
		return ""
	}

	guideFocus := onListPanel
	if m.focus == onPreviewer {
		guideFocus = onPreviewer
	}

	focused := m.keys.forFocus(guideFocus)
	full := m.help.FullHelpView(focused.FullHelp())

	return m.styles.Help.
		Padding(1, 2).
		Width(m.width).
		Render(full)
}
