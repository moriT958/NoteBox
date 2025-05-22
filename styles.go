package main

import (
	"github.com/charmbracelet/lipgloss"
)

/* Styles */
var (
	appStyle      = lipgloss.NewStyle().Padding(1, 2)
	appTitleStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: borderColor}).
			Align(lipgloss.Center).
			Padding(0, 1)
)

/* Border Style */

const (
	cursorColor       = "#f2d5cf"
	borderActiveColor = "#babbf1"
	borderColor       = "#737994"
)

func borderStyle(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true, true, true, true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: borderActiveColor}).
			Padding(0, 0, 0, 1)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: borderColor}).
		Padding(0, 0, 0, 1)
}
