package logger

import (
	"fmt"
	"log/slog"
	"notebox/internal/config"
	"os"
	"path/filepath"
)

func Set() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filename := filepath.Join(home, config.AppDirName, config.LogFileName)

	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(fp, nil))
	slog.SetDefault(logger)

	return nil
}
