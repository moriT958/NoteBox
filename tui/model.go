package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"notebox/models"
	"strings"
)

type focusedArea int

const (
	focusListArea focusedArea = iota
	focusPreviewArea
)

type mode int

const (
	createMode mode = iota
	navigateMode
)

// Message types
type errMsg struct{ error }
type editorFinishedMsg struct{ error }

func initModel(notes []*models.Note) *model {
	return defaultModelConfig(notes)
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("NoteBox")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case editorFinishedMsg:
		m.previewerRender()

	case tea.KeyMsg:
		cmd = m.handleKeyInput(msg, cmd)
	}

	return m, cmd
}

func (m *model) handleKeyInput(msg tea.KeyMsg, cmd tea.Cmd) tea.Cmd {
	var cmds []tea.Cmd

	if m.input.Focused() {
		cmd = m.handleInputModal(msg)
		cmds = append(cmds, cmd)

	} else if m.warnModal.open {
		cmd = m.handleWarnModal(msg)
		cmds = append(cmds, cmd)

	} else if m.focus == focusListArea {
		cmd = m.handleListPanel(msg)
		cmds = append(cmds, cmd)

	} else if m.focus == focusPreviewArea {
		cmd = m.handlePreviewer(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	final := strings.Builder{}

	listPanel := m.listPanelRender()
	final.WriteString(listPanel)

	if m.input.Focused() {
		prompt := "Enter note name:\n\n" + m.input.View()
		final.WriteString("\n" + docStyle.Render(prompt) + "\n")
		return final.String()
	}

	if m.warnModal.open {
		final.WriteString("\n" + docStyle.Render(m.warnModal.message) + "\n")
		return final.String()
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, final.String(), m.viewport.View()+m.helpView())
}

func (m model) helpView() string {
	if m.focus == focusListArea {
		return helpStyle.Render("\n ↑/↓: Navigate • e: Edit • ctrl+l: focus-preview • q: Quit")
	} else if m.focus == focusPreviewArea {
		return helpStyle.Render("\n  ↑/↓: Navigate • ctrl+h: focus-list • q: Quit\n")
	}
	return ""
}

func (m *model) deleteNoteWarn() {
	m.warnModal = warnModal{
		open:    true,
		message: "Are you sure you want to completely delete?(y/n)",
	}
}

func (m model) getNoteFromCursor() *models.Note {
	id := m.listPanel.notes[m.listPanel.cursor].ID
	note, err := models.GetRepository().FindByID(id)
	if err != nil {
		return nil
	}
	return note
}
