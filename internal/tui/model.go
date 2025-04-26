package tui

import (
	"log/slog"

	"NoteBox.tmp/internal/config"
	listpanel "NoteBox.tmp/internal/tui/listPanel"
	"NoteBox.tmp/internal/tui/previewer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusListPanel int = iota
	focusPreviewer
)

type Model struct {
	focus     int
	listPanel listpanel.Model
	previewer previewer.Model
}

func New(cfg *config.Config) *Model {

	return &Model{
		focus:     focusListPanel,
		listPanel: *listpanel.New(cfg.Volume),
		previewer: *previewer.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("NoteBox")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+h", "left":
			m.focus = focusListPanel

		case "ctrl+l", "right":
			m.focus = focusPreviewer
		}

		if m.focus == focusListPanel {
			m.listPanel, cmd = m.listPanel.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.focus == focusPreviewer {
			m.previewer, cmd = m.previewer.Update(msg)
			cmds = append(cmds, cmd)
		}

	case error:
		slog.Error(msg.Error())
		return m, tea.Quit

	default:
		m.listPanel, cmd = m.listPanel.Update(msg)
		cmds = append(cmds, cmd)
		m.previewer, cmd = m.previewer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Left, m.listPanel.View(), m.previewer.View())
}
