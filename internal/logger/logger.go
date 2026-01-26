package logger

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"notebox/internal/config"
	"notebox/internal/utils"
	"os"
	"path/filepath"
)

func Set() error {
	filename := filepath.Join(utils.HomeDir(), config.AppDirName, config.LogFileName)

	if _, err := os.Stat(filename); errors.Is(err, fs.ErrNotExist) {
		if _, err = os.Create(filename); err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
	}

	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	// defer fp.Close()

	logger := slog.New(slog.NewTextHandler(fp, nil))
	slog.SetDefault(logger)

	return nil
}
