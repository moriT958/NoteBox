package main

import (
	"context"
	"fmt"
	"log/slog"
	"notebox/internal/cli"
	"notebox/internal/config"
	"notebox/internal/logger"
	"notebox/internal/tui"
	"notebox/internal/utils"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		m, err := tui.NewModel()
		if err != nil {
			slog.Error("failed to initialize bubbletea model", "error", err)
			os.Exit(1)
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			slog.Error("failed to run bubbletea app", "error", err)
			os.Exit(1)
		}
	} else {
		os.Exit(cli.InitCommands(context.Background()))
	}
}

func init() {
	// ensure .notebox dir exits.
	noteboxPath := filepath.Join(utils.HomeDir(), config.AppDirName)
	if err := os.MkdirAll(noteboxPath, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "failed to make notebox dir:", err)
		os.Exit(1)
	}

	// set logger
	if err := logger.Set(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to set logger:", err)
		os.Exit(1)
	}
}
