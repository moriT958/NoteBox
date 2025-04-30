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
	var fp *os.File
	if _, err := os.Stat("notebox.log"); errors.Is(err, fs.ErrNotExist) {
		fp, err = os.Create("notebox.log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log file: %v\n", err)
			os.Exit(1)
		}
		defer fp.Close()
	}
	logger := slog.New(slog.NewTextHandler(fp, nil))
	slog.SetDefault(logger)

	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
