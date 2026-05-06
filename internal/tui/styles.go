package tui

import "charm.land/lipgloss/v2"

type styles struct {
	main              lipgloss.Style
	header            lipgloss.Style
	borderActive      lipgloss.Style
	borderPassive     lipgloss.Style
	cursorColor       lipgloss.Style
	modalConfirmColor lipgloss.Style
	modalCalcelColor  lipgloss.Style
}

type theme bool

const (
	darkTheme  theme = true
	lightTheme theme = false
)

func defaultStyles() *styles {
	s := new(styles)

	s.main = lipgloss.NewStyle()
	s.header = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder(), false, false, true, false).
		BorderForeground(lipgloss.LightDark(true)(lipgloss.Color("#F793FF"), lipgloss.Color("#737994")))

	s.borderActive = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true, true, true, true).
		BorderForeground(lipgloss.LightDark(true)(lipgloss.Color("#F793FF"), lipgloss.Color("#babbf1")))

	s.borderPassive = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true, true, true, true).
		BorderForeground(lipgloss.LightDark(true)(lipgloss.Color("#F793FF"), lipgloss.Color("#737994")))

	s.cursorColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#f2d5cf"))
	s.modalConfirmColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#99d1db"))
	s.modalCalcelColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#ea999c"))

	return s
}

func adjustSize(width, height int) func(strs ...string) string {
	return lipgloss.NewStyle().Height(height).Width(width).Render
}
