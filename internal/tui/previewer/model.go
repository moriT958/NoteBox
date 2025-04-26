package previewer

import (
	"fmt"

	listpanel "NoteBox.tmp/internal/tui/listPanel"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	viewport viewport.Model
	renderer *glamour.TermRenderer
}

func New() *Model {
	vp := viewport.New(0, 0)
	vp.SetContent("{{ No note selected }}")

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	return &Model{
		viewport: vp,
		renderer: renderer,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight

	case listpanel.NoteMsg:
		content, err := msg.Content()
		if err != nil {
			return m, func() tea.Msg { return err }
		}
		rendered, err := m.renderer.Render(content)
		if err != nil {
			return m, func() tea.Msg { return err }
		}
		m.viewport.SetContent(rendered)

	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}
