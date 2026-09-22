package main

import (
	"context"
	"fmt"
	"log/slog"
	"notebox/internal/cli"
	"notebox/internal/config"
	"notebox/internal/database"
	"notebox/internal/logger"
	"notebox/internal/notedeprecated"
	"notebox/internal/tui"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// initialize DB and run migrations; shared by the TUI and every CLI subcommand
	db, err := database.NewSQLiteDB()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	boxRepo := database.NewBoxRepository(db)

	if len(os.Args) < 2 {
		reg, err := notedeprecated.NewFSNotifyRegisterer()
		if err != nil {
			slog.Error("failed to initialize fsnotify watcher", "error", err)
			os.Exit(1)
		}
		defer reg.Close()

		m, err := tui.NewModel(reg, boxRepo)
		if err != nil {
			slog.Error("failed to initialize bubbletea model", "error", err)
			os.Exit(1)
		}

		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			slog.Error("failed to run bubbletea app", "error", err)
			os.Exit(1)
		}
	} else {
		os.Exit(cli.InitCommands(context.Background(), boxRepo))
	}
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get home dir: %v", err)
		os.Exit(1)
	}

	// ensure .notebox dir exits.
	noteboxPath := filepath.Join(home, config.AppDirName)
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
