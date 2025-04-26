package main

import (
	"log/slog"
	"os"

	"NoteBox.tmp/internal/config"
	"NoteBox.tmp/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	// Setup logger
	fp, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer fp.Close()
	logger := slog.New(slog.NewTextHandler(fp, &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	config, err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error())
	}

	m := tui.New(config)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
