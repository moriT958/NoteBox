package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"notebox/internal/utils"
	"os"
	"path/filepath"
	"sync"
)

var (
	// Set by goreleaser
	CurrentVersion string = "dev"
)

var (
	instance *Config
	once     sync.Once
	initErr  error
)

type Config struct {
	Editor       string `json:"editor"`
	NotesDir     string `json:"notesdir"`
	DummyNoteDir string `json:"-"`
}

func GetConfig() (*Config, error) {
	once.Do(func() {
		instance, initErr = loadConfig()
	})
	return instance, initErr
}

func loadConfig() (*Config, error) {
	filename := filepath.Join(utils.HomeDir(), AppDirName, ConfigFileName)

	// Create default setting file if not exist.
	if _, err := os.Stat(filename); errors.Is(err, fs.ErrNotExist) {
		fp, createErr := os.Create(filename)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create config file: %v", createErr)
		}
		defer fp.Close()

		enc := json.NewEncoder(fp)
		enc.SetIndent("", "  ")
		if err := enc.Encode(defaultConfig()); err != nil {
			return nil, fmt.Errorf("failed to encode config file: %v", err)
		}
	}

	// Read setting file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	cfg := new(Config)
	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	cfg.DummyNoteDir = filepath.Join(utils.HomeDir(), AppDirName, DummyFileName)

	if err := ensureDirectoriesAndFiles(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Editor:   DefaultEditor,
		NotesDir: filepath.Join(utils.HomeDir(), AppDirName, NotesDirName),
	}
}

func ensureDirectoriesAndFiles(cfg *Config) error {
	if err := os.MkdirAll(cfg.NotesDir, 0755); err != nil {
		return fmt.Errorf("failed to create notes dir: %v", err)
	}

	fp, err := os.Create(cfg.DummyNoteDir)
	if err != nil {
		return fmt.Errorf("failed to create dummy note: %v", err)
	}
	defer fp.Close()
	fmt.Fprint(fp, DummyNoteContent)

	return nil
}
