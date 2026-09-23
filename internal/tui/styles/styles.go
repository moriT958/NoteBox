package styles

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type TabStyles struct {
	Active          lipgloss.Style
	Inactive        lipgloss.Style
	ActivePreview   lipgloss.Style
	InactivePreview lipgloss.Style
}

type Styles struct {
	Header lipgloss.Style
	Help   lipgloss.Style
	Error  lipgloss.Style

	// Panel borders and the frame color used when drawing borders by hand
	// (the preview's tab bar doubles as its top border).
	BorderFocused lipgloss.Style
	BorderBlurred lipgloss.Style
	FrameFocused  lipgloss.Style
	FrameBlurred  lipgloss.Style

	List struct {
		Cursor       lipgloss.Style
		Path         lipgloss.Style
		SelectedPath lipgloss.Style
	}

	Tab struct {
		Focused TabStyles
		Blurred TabStyles
	}

	Dialog struct {
		Frame   lipgloss.Style
		Title   lipgloss.Style
		Cursor  lipgloss.Style
		Missing lipgloss.Style
		Confirm lipgloss.Style
		Cancel  lipgloss.Style
		Error   lipgloss.Style
	}
}

// New builds the styles for the given theme name ("dark" or "light").
func New(theme string) (*Styles, error) {
	var isDark bool
	switch theme {
	case "dark":
		isDark = true
	case "light":
		isDark = false
	default:
		return nil, fmt.Errorf("unknown color theme: %q", theme)
	}
	pick := lipgloss.LightDark(isDark)

	primary := pick(lipgloss.Color("#b89988"), lipgloss.Color("#737994"))
	active := pick(lipgloss.Color("#fe640b"), lipgloss.Color("#babbf1"))
	cursor := pick(lipgloss.Color("#dc8a78"), lipgloss.Color("#f2d5cf"))
	cursorMuted := pick(lipgloss.Color("#b58c7e"), lipgloss.Color("#c7a89e"))
	muted := pick(lipgloss.Color("#8c8fa1"), lipgloss.Color("#51576d"))
	tabBarBg := lipgloss.Color("#232634")
	buttonFg := pick(lipgloss.Color("#eff1f5"), lipgloss.Color("#414559"))
	confirmBg := pick(lipgloss.Color("#40a02b"), lipgloss.Color("#99d1db"))
	cancelBg := pick(lipgloss.Color("#d20f39"), lipgloss.Color("#ea999c"))

	s := &Styles{}

	s.Header = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder(), false, false, true, false).
		BorderForeground(primary).
		Foreground(active).
		Bold(true).
		Align(lipgloss.Center)
	s.Help = lipgloss.NewStyle().PaddingLeft(1)
	s.Error = lipgloss.NewStyle().PaddingLeft(1).Foreground(cancelBg)

	s.BorderFocused = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(active)
	s.BorderBlurred = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(primary)
	s.FrameFocused = lipgloss.NewStyle().Foreground(active)
	s.FrameBlurred = lipgloss.NewStyle().Foreground(primary)

	s.List.Cursor = lipgloss.NewStyle().Foreground(cursor)
	s.List.Path = lipgloss.NewStyle().Foreground(muted)
	s.List.SelectedPath = lipgloss.NewStyle().Foreground(cursorMuted)

	s.Tab.Focused = TabStyles{
		Active:          lipgloss.NewStyle().Foreground(active).Background(tabBarBg).Bold(true),
		Inactive:        lipgloss.NewStyle().Foreground(active),
		ActivePreview:   lipgloss.NewStyle().Foreground(active).Background(tabBarBg).Bold(true).Italic(true),
		InactivePreview: lipgloss.NewStyle().Foreground(active).Italic(true),
	}
	s.Tab.Blurred = TabStyles{
		Active:          lipgloss.NewStyle().Foreground(primary).Bold(true),
		Inactive:        lipgloss.NewStyle().Foreground(primary),
		ActivePreview:   lipgloss.NewStyle().Foreground(primary).Bold(true).Italic(true),
		InactivePreview: lipgloss.NewStyle().Foreground(primary).Italic(true),
	}

	s.Dialog.Frame = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(active).
		Padding(1, 2)
	s.Dialog.Title = lipgloss.NewStyle().Foreground(active).Bold(true)
	s.Dialog.Cursor = lipgloss.NewStyle().Foreground(cursor)
	s.Dialog.Missing = lipgloss.NewStyle().Strikethrough(true)
	s.Dialog.Confirm = lipgloss.NewStyle().Foreground(buttonFg).Background(confirmBg)
	s.Dialog.Cancel = lipgloss.NewStyle().Foreground(buttonFg).Background(cancelBg)
	s.Dialog.Error = lipgloss.NewStyle().Foreground(cancelBg)

	return s, nil
}
