package tui

import (
	"notebox/models"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func defaultModelConfig(notes []*models.Note) *model {
	height := 45
	width := 120
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	const glamourGutter = 2
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		renderer = nil
	}

	ti := textinput.New()
	ti.Placeholder = "New note name..."
	ti.Prompt = "$ "
	ti.CharLimit = 50
	ti.Width = 30

	m := &model{
		listPanel: listPanel{
			height: 45,
			width:  10,
			notes:  notes,
			cursor: 0,
		},
		focus: focusListArea,
		mode:  navigateMode,
		warnModal: warnModal{
			open:    false,
			message: "",
		},
		viewport: vp,
		renderer: renderer,
		input:    ti,
	}

	m.viewport.SetContent("(No note selected)")
	m.viewport.GotoTop()

	return m
}
