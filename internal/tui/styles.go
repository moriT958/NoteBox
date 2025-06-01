package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	main              lipgloss.Style
	header            lipgloss.Style
	borderActive      lipgloss.Style
	borderPassive     lipgloss.Style
	cursorColor       lipgloss.Style
	modalConfirmColor lipgloss.Style
	modalCalcelColor  lipgloss.Style
}

func defaultStyles() *styles {
	s := new(styles)

	s.main = lipgloss.NewStyle().Padding(1, 2)
	s.header = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder(), false, false, true, false).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#737994"})
	s.borderActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#babbf1"})
	s.borderPassive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#737994"})
	s.cursorColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#f2d5cf"))
	s.modalConfirmColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#99d1db"))
	s.modalCalcelColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#ea999c"))

	return s
}

func adjustSize(width, height int) func(strs ...string) string {
	return lipgloss.NewStyle().Height(height).Width(width).Render
}
