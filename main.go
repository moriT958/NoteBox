package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"notebox/cli"
	"notebox/config"
	"notebox/logger"
	"notebox/utils"

	tea "github.com/charmbracelet/bubbletea"
)

/* MAIN */

func main() {
	slog.Info("hello, world!")
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

func init() {
	// ensure .notebox dir exits.
	noteboxPath := filepath.Join(utils.HomeDir(), ".notebox")
	if err := os.MkdirAll(noteboxPath, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "failed to make notebox dir:", err)
		os.Exit(1)
	}

	// load config
	if err := config.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "failed load config:", err)
		os.Exit(1)
	}

	// set logger
	if err := logger.Set(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to set logger:", err)
		os.Exit(1)
	}
}
