package styles

import (
	"errors"
	"notebox/internal/config"

	"charm.land/lipgloss/v2"
)

type Theme int

const (
	DarkTheme Theme = iota
	LightTheme
)

type TabBarStyles struct {
	Active          lipgloss.Style
	Inactive        lipgloss.Style
	ActivePreview   lipgloss.Style
	InactivePreview lipgloss.Style
}

type Style struct {
	Main          lipgloss.Style
	BorderActive  lipgloss.Style
	BorderPassive lipgloss.Style
	Cursor        lipgloss.Style
	Help          lipgloss.Style
	Header        lipgloss.Style

	ActiveColor  lipgloss.Style
	PassiveColor lipgloss.Style
	NotePath     lipgloss.Style
	CursorPath   lipgloss.Style

	TabBarFocused    TabBarStyles
	TabBarUnforcused TabBarStyles

	Modal ModalStyle
}

type ModalStyle struct {
	Centered lipgloss.Style
	Fuzzy    lipgloss.Style
	Confirm  lipgloss.Style
	Cancel   lipgloss.Style
}

func New(theme Theme) *Style {
	isDark := theme == DarkTheme

	primary := lipgloss.LightDark(isDark)(lipgloss.Color("#b89988"), lipgloss.Color("#737994"))
	active := lipgloss.LightDark(isDark)(lipgloss.Color("#fe640b"), lipgloss.Color("#babbf1"))
	cursor := lipgloss.LightDark(isDark)(lipgloss.Color("#dc8a78"), lipgloss.Color("#f2d5cf"))
	cursorMuted := lipgloss.LightDark(isDark)(lipgloss.Color("#b58c7e"), lipgloss.Color("#c7a89e"))
	muted := lipgloss.LightDark(isDark)(lipgloss.Color("#8c8fa1"), lipgloss.Color("#51576d"))

	tabBarBg := lipgloss.Color("#232634")

	confirmFg := lipgloss.LightDark(isDark)(lipgloss.Color("#eff1f5"), lipgloss.Color("#414559"))
	confirmBg := lipgloss.LightDark(isDark)(lipgloss.Color("#40a02b"), lipgloss.Color("#99d1db"))
	cancelFg := lipgloss.LightDark(isDark)(lipgloss.Color("#eff1f5"), lipgloss.Color("#414559"))
	cancelBg := lipgloss.LightDark(isDark)(lipgloss.Color("#d20f39"), lipgloss.Color("#ea999c"))

	return &Style{
		Main: lipgloss.NewStyle(),

		ActiveColor:  lipgloss.NewStyle().Foreground(active),
		PassiveColor: lipgloss.NewStyle().Foreground(primary),
		NotePath:     lipgloss.NewStyle().Foreground(muted),
		CursorPath:   lipgloss.NewStyle().Foreground(cursorMuted),

		BorderActive: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), true, true, true, true).
			BorderForeground(active),

		BorderPassive: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), true, true, true, true).
			BorderForeground(primary),

		Cursor: lipgloss.NewStyle().Foreground(cursor),

		Help: lipgloss.NewStyle().
			Align(lipgloss.Left).
			PaddingLeft(1),

		Header: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(primary).
			Foreground(active).
			Bold(true).
			Align(lipgloss.Center),

		TabBarFocused: TabBarStyles{
			Active:          lipgloss.NewStyle().Foreground(active).Background(tabBarBg).Bold(true),
			Inactive:        lipgloss.NewStyle().Foreground(active),
			ActivePreview:   lipgloss.NewStyle().Foreground(active).Background(tabBarBg).Bold(true).Italic(true),
			InactivePreview: lipgloss.NewStyle().Foreground(active).Italic(true),
		},

		TabBarUnforcused: TabBarStyles{
			Active:          lipgloss.NewStyle().Foreground(primary).Bold(true),
			Inactive:        lipgloss.NewStyle().Foreground(primary),
			ActivePreview:   lipgloss.NewStyle().Foreground(primary).Bold(true).Italic(true),
			InactivePreview: lipgloss.NewStyle().Foreground(primary).Italic(true),
		},

		Modal: ModalStyle{
			Centered: lipgloss.NewStyle().
				Align(lipgloss.Center, lipgloss.Center),

			Fuzzy: lipgloss.NewStyle().
				Padding(1, 2),

			Confirm: lipgloss.NewStyle().
				Foreground(confirmFg).
				Background(confirmBg),

			Cancel: lipgloss.NewStyle().
				Foreground(cancelFg).
				Background(cancelBg),
		},
	}
}

func (s *Style) Sized(width, height int) lipgloss.Style {
	return s.Main.Width(width).Height(height)
}

func GetColorTheme(cfg *config.Config) (Theme, error) {
	switch cfg.Theme {
	case "dark":
		return DarkTheme, nil
	case "light":
		return LightTheme, nil
	default:
		return 0, errors.New("unknown color theme")
	}
}
