package listpanel

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	volume string
	notes  []Note
	list   list.Model
}

func New(volume string) *Model {
	notes := getAllNoteFiles(volume)
	items := make([]list.Item, len(notes))
	for i, n := range notes {
		items[i] = n
	}
	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = "My Notes"

	m := &Model{
		volume: volume,
		notes:  notes,
		list:   list,
	}
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.String() {
		default:
			cmd = func() tea.Msg { return NoteMsg{m.notes[m.list.GlobalIndex()]} }
			cmds = append(cmds, cmd)
		}

	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}
