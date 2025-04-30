package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* Note */

type note struct {
	title   string
	content string
}

var notes []note = []note{
	{title: "sample", content: "sample0"},
	{title: "sample", content: "sample1"},
	{title: "sample", content: "sample2"},
	{title: "sample", content: "sample3"},
	{title: "sample", content: "sample4"},
	{title: "sample", content: "sample5"},
}

/* Main Model */

type model struct {
	height, width int
	focus         int
	listPanel     listPanelModel
	previewer     previewerModel
	typingModal   typingModal
}

const (
	focusListPanel int = iota
	focusPreviewer
	focusTypingModal
)

func newModel() *model {

	tm := &typingModal{
		open:  false,
		input: textinput.New(),
	}

	lp := &listPanelModel{
		cursor: 0,
		notes:  notes,
	}

	pm := &previewerModel{
		vp: viewport.New(0, 0),
	}

	m := &model{
		height:      0,
		width:       0,
		focus:       focusListPanel,
		listPanel:   *lp,
		previewer:   *pm,
		typingModal: *tm,
	}
	return m
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	// init listPanel
	cmds = append(cmds, m.listPanel.init())
	// init previewer
	cmds = append(cmds, m.previewer.init())
	// init typingModal
	cmds = append(cmds, m.typingModal.init())

	cmds = append(cmds, tea.SetWindowTitle("NoteBox"))
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "ctrl+h":
			m.focus = focusListPanel
		case "ctrl+l":
			m.focus = focusPreviewer
		}

		switch m.focus {
		case focusListPanel:
			m.listPanel, cmd = m.listPanel.update(msg)
			cmds = append(cmds, cmd)
		case focusPreviewer:
			m.previewer, cmd = m.previewer.update(msg)
			cmds = append(cmds, cmd)
		case focusTypingModal:
			m.typingModal, cmd = m.typingModal.update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.previewer, cmd = m.previewer.update(msg)
		cmds = append(cmds, cmd)
		m.typingModal, cmd = m.typingModal.update(msg)
		cmds = append(cmds, cmd)
	case typingModalMsg:
		if msg.isOpen {
			m.focus = focusTypingModal
		} else {
			m.focus = focusListPanel
		}
		m.typingModal, cmd = m.typingModal.update(msg)
		cmds = append(cmds, cmd)
	case renderPreviewMsg:
		m.previewer, cmd = m.previewer.update(msg)
		cmds = append(cmds, cmd)
	case createNewNoteMsg:
		m.listPanel, cmd = m.listPanel.update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		"NoteBox",
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.listPanel.view(),
			m.previewer.view(),
			m.typingModal.view()))
}

/* List Panel Model */

type listPanelModel struct {
	cursor int
	notes  []note
}

func (m listPanelModel) init() tea.Cmd {
	return func() tea.Msg {
		return renderPreviewMsg{m.notes[m.cursor].content}
	}
}

type typingModalMsg struct{ isOpen bool }

func (m listPanelModel) update(msg tea.Msg) (listPanelModel, tea.Cmd) {
	var (
		cmd tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.notes)-1 {
				m.cursor++
			}
			cmd = func() tea.Msg { return renderPreviewMsg{m.notes[m.cursor].content} }
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			cmd = func() tea.Msg { return renderPreviewMsg{m.notes[m.cursor].content} }
		case "n":
			cmd = func() tea.Msg { return typingModalMsg{true} }
		}
	case createNewNoteMsg:
		newNoteContent := fmt.Sprintf("# %s\n\n", msg.title)
		m.notes = append(m.notes, note{msg.title, newNoteContent})
		m.cursor = len(m.notes) - 1
		cmd = func() tea.Msg { return renderPreviewMsg{m.notes[m.cursor].content} }
	}
	return m, cmd
}

func (m listPanelModel) view() string {
	view := strings.Builder{}
	for i, n := range m.notes {
		if i == m.cursor {
			view.WriteString(">" + n.title)
		} else {
			view.WriteString(n.title)
		}
		view.WriteString("\n")
	}
	return view.String()
}

/* Previewer Model */

type renderPreviewMsg struct{ content string }

type previewerModel struct {
	vp viewport.Model
}

func (m previewerModel) init() tea.Cmd {
	return nil
}

func (m previewerModel) update(msg tea.Msg) (previewerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.vp.Height = msg.Height * 2 / 3
		m.vp.Width = msg.Width * 2 / 3
	case renderPreviewMsg:
		m.vp.SetContent(msg.content)
	}

	m.vp, cmd = m.vp.Update(msg)

	return m, cmd
}

func (m previewerModel) view() string {
	view := strings.Builder{}
	view.WriteString(m.vp.View())
	return view.String()
}

/* Typing Modal Model */

type createNewNoteMsg struct{ title string }

type typingModal struct {
	open  bool
	input textinput.Model
}

func (m typingModal) init() tea.Cmd {
	return textinput.Blink
}

func (m typingModal) update(msg tea.Msg) (typingModal, tea.Cmd) {
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
			createNewNoteCmd := func() tea.Msg {
				title := m.input.Value()
				return createNewNoteMsg{title}
			}
			closeModalCmd := func() tea.Msg {
				return typingModalMsg{false}
			}
			cmds = append(cmds, createNewNoteCmd, closeModalCmd)
		case "esc", "ctrl+c":
			m.open = false
			cmds = append(cmds, func() tea.Msg { return typingModalMsg{false} })
		}
	case typingModalMsg:
		m.input.Reset()
		m.open = msg.isOpen
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m typingModal) view() string {
	var view strings.Builder
	if m.open {
		view.WriteString(m.input.View())
	}
	return view.String()
}
