package styles

import (
	"notebox/internal/mermaid"

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
	Info   lipgloss.Style

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
		// Hover and HoverPath draw the title and the path of the note under
		// the mouse pointer.
		Hover     lipgloss.Style
		HoverPath lipgloss.Style
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

	// Selection marks the text selected with the mouse in the preview.
	Selection lipgloss.Style

	// Mermaid colors the diagrams drawn in the preview.
	Mermaid mermaid.Styles
}

// New builds the styles from the colors of a palette.
func New(p Palette) *Styles {
	primary := lipgloss.Color(p.Primary)
	active := lipgloss.Color(p.Accent)
	cursor := lipgloss.Color(p.Cursor)
	cursorMuted := lipgloss.Color(p.CursorMuted)
	muted := lipgloss.Color(p.Muted)
	hoverBg := lipgloss.Color(p.ListHoverBg)
	tabBarBg := lipgloss.Color(p.TabBarBg)
	buttonFg := lipgloss.Color(p.ButtonFg)
	confirmBg := lipgloss.Color(p.Success)
	cancelBg := lipgloss.Color(p.Danger)

	s := &Styles{}

	s.Header = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder(), false, false, true, false).
		BorderForeground(primary).
		Foreground(active).
		Bold(true).
		Align(lipgloss.Center)
	s.Help = lipgloss.NewStyle().PaddingLeft(1)
	s.Error = lipgloss.NewStyle().PaddingLeft(1).Foreground(cancelBg)
	s.Info = lipgloss.NewStyle().PaddingLeft(1).Foreground(confirmBg)

	s.BorderFocused = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(active)
	s.BorderBlurred = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(primary)
	s.FrameFocused = lipgloss.NewStyle().Foreground(active)
	s.FrameBlurred = lipgloss.NewStyle().Foreground(primary)

	s.List.Cursor = lipgloss.NewStyle().Foreground(cursor)
	s.List.Path = lipgloss.NewStyle().Foreground(muted)
	s.List.SelectedPath = lipgloss.NewStyle().Foreground(cursorMuted)
	s.List.Hover = lipgloss.NewStyle().Background(hoverBg)
	s.List.HoverPath = lipgloss.NewStyle().Foreground(muted).Background(hoverBg)

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

	s.Selection = lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.SelectionFg)).
		Background(lipgloss.Color(p.SelectionBg))

	s.Mermaid = mermaid.Styles{
		Border:    lipgloss.Color(p.MermaidBorder),
		Text:      lipgloss.Color(p.MermaidText),
		Edge:      lipgloss.Color(p.MermaidEdge),
		EdgeLabel: lipgloss.Color(p.MermaidEdgeLabel),
	}

	return s
}
