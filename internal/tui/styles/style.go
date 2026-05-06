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

type Style struct {
	Main          lipgloss.Style
	Header        lipgloss.Style
	BorderActive  lipgloss.Style
	BorderPassive lipgloss.Style
	Cursor        lipgloss.Style

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

	primary := lipgloss.LightDark(isDark)(lipgloss.Color("#b89988"), lipgloss.Color("#F793FF"))
	active := lipgloss.LightDark(isDark)(lipgloss.Color("#fe640b"), lipgloss.Color("#F793FF"))
	cursor := lipgloss.LightDark(isDark)(lipgloss.Color("#dc8a78"), lipgloss.Color("#f2d5cf"))

	confirmFg := lipgloss.LightDark(isDark)(lipgloss.Color("#eff1f5"), lipgloss.Color("#414559"))
	confirmBg := lipgloss.LightDark(isDark)(lipgloss.Color("#40a02b"), lipgloss.Color("#99d1db"))
	cancelFg := lipgloss.LightDark(isDark)(lipgloss.Color("#eff1f5"), lipgloss.Color("#414559"))
	cancelBg := lipgloss.LightDark(isDark)(lipgloss.Color("#d20f39"), lipgloss.Color("#ea999c"))

	return &Style{
		Main: lipgloss.NewStyle(),

		Header: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			BorderForeground(primary),

		BorderActive: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), true, true, true, true).
			BorderForeground(active),

		BorderPassive: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), true, true, true, true).
			BorderForeground(primary),

		Cursor: lipgloss.NewStyle().Foreground(cursor),

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
