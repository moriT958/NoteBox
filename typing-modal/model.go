package typingmodal

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* Typing Modal Model */

type Model struct {
	width, height int
	open          bool
	input         textinput.Model
}

func New() Model {
	m := Model{
		width:  60,
		height: 7,
		open:   false,
		input:  textinput.New(),
	}
	return m
}

func (m Model) GetSize() (int, int) {
	return m.width, m.height
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.input.Placeholder = "Enter note name..."
		m.input.Focus()
		m.input.CharLimit = 50
		m.input.Width = msg.Width / 3
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.open = false
			cmds = append(cmds,
				m.inputCmd(),
				ToggleModalCmd(false))
		case "esc", "ctrl+c":
			m.open = false
			cmds = append(cmds, ToggleModalCmd(false))
		}
	case TypingModalMsg:
		m.input.Reset()
		m.open = bool(msg)
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.open {
		confirm := ModalConfirm.Render(" (" + "enter" + ") Create ")
		cancel := ModalCancel.Render(" (" + "ctrl+c" + ") Cancel ")

		tip := confirm +
			lipgloss.NewStyle().Render("           ") +
			cancel

		return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
			Render("\n" + m.input.View() + "\n\n" + tip)
	}
	return ""
}
