package main

import (
	"context"
	"fmt"
	"log/slog"
	"notebox/internal/cli"
	"notebox/internal/config"
	"notebox/internal/core/box"
	"notebox/internal/database"
	"notebox/internal/logger"
	"notebox/internal/tui"
	"notebox/internal/watcher"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// open the DB and run migrations; shared by the TUI and every CLI subcommand
	store, err := database.NewSQLiteBoxStore()
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get home dir", "error", err)
		os.Exit(1)
	}
	// new boxes are created under the app dir unless a path is given
	boxes := box.NewBoxService(filepath.Join(home, config.AppDirName), store)

	if len(os.Args) >= 2 {
		os.Exit(cli.InitCommands(context.Background(), boxes))
	}

	cfg, err := config.GetConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	w := watcher.NewWatcher()
	defer w.Close()

	m, err := tui.New(cfg, boxes, w)
	if err != nil {
		slog.Error("failed to initialize tui", "error", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		slog.Error("failed to run tui", "error", err)
		os.Exit(1)
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
