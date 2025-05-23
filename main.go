package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"NoteBox.tmp/cli"
	"NoteBox.tmp/config"
	tea "github.com/charmbracelet/bubbletea"
)

/* File Operations */

/* MAIN */

func main() {

	if _, err := os.Stat(config.LogfilePath); errors.Is(err, fs.ErrNotExist) {
		if _, err = os.Create(config.LogfilePath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log file: %v\n", err)
			os.Exit(1)
		}
	}

	fp, err := os.OpenFile(config.LogfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer fp.Close()

	logger := slog.New(slog.NewTextHandler(fp, nil))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
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
	} else {
		os.Exit(cli.InitCommands(context.Background()))
	}
}
