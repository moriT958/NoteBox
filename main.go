package main

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/fs"
	"log/slog"
	"os"
)

/* File Operations */

/* MAIN */

func main() {

	if _, err := os.Stat("notebox.log"); errors.Is(err, fs.ErrNotExist) {
		if _, err = os.Create("notebox.log"); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log file: %v\n", err)
			os.Exit(1)
		}
	}

	fp, err := os.OpenFile("notebox.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer fp.Close()

	logger := slog.New(slog.NewTextHandler(fp, nil))
	slog.SetDefault(logger)

	m, err := newModel()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to initialize bubbletea model: %v", err))
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error(fmt.Sprintf("failed to run bubbletea app: %v", err))
		os.Exit(1)
	}
}
