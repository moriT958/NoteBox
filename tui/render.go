package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) listPanelRender() string {
	s := strings.Builder{}
	s.WriteString("📓 Your Notes 📓\n\n")
	for i := range len(m.listPanel.notes) {
		if m.listPanel.cursor == i {
			s.WriteString(lipgloss.NewStyle().
				Reverse(true).
				Render(m.listPanel.notes[i].Title))
		} else {
			s.WriteString(m.listPanel.notes[i].Title)
			s.WriteString(" ")
		}
		s.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Margin(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).Render(s.String())
}

func (m *model) previewerRender() {
	note := m.listPanel.notes[m.listPanel.cursor]
	content, err := os.ReadFile(note.GetFilePath())
	if err != nil {
		m.Update(errMsg{err})
	}
	str, err := m.renderer.Render(string(content))
	if err != nil {
		m.Update(errMsg{err})
	}
	m.viewport.SetContent(str)
	m.viewport.GotoTop()
}
