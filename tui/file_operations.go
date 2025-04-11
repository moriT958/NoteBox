package tui

import (
	"notebox/config"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) openFileWithEditor(file string) tea.Cmd {
	editor := config.Editor()
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	parts := strings.Fields(editor)
	cmd := parts[0]
	args := append(parts[1:], file)

	c := exec.Command(cmd, args...)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}
