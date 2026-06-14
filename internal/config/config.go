package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	Theme        string `json:"theme"`
	DummyNoteDir string `json:"-"`
}

func GetConfig() (*Config, error) {
	once.Do(func() {
		instance, initErr = loadConfig()
	})
	return instance, initErr
}

func loadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(home, AppDirName, ConfigFileName)

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

	cfg.DummyNoteDir = filepath.Join(home, AppDirName, DummyFileName)

	if err := ensureDirectoriesAndFiles(cfg); err != nil {
		return nil, err
	}

	cfg.Theme = strings.ToLower(cfg.Theme)

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Editor: DefaultEditor,
		Theme:  defaultTheme,
	}
}

func ensureDirectoriesAndFiles(cfg *Config) error {
	fp, err := os.Create(cfg.DummyNoteDir)
	if err != nil {
		return fmt.Errorf("failed to create dummy note: %v", err)
	}
	defer fp.Close()
	fmt.Fprint(fp, DummyNoteContent)

	return nil
}

// LoadLastBoxID reads the last used box ID from the state file.
// Returns 0 (no preference) when the file does not exist or contains invalid content.
func LoadLastBoxID() (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(filepath.Join(home, AppDirName, StateFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	id, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, nil
	}
	return id, nil
}

// SaveLastBoxID writes the given box ID to the state file asynchronously.
func SaveLastBoxID(id int) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, AppDirName, StateFileName), []byte(strconv.Itoa(id)), 0644)
}

// DefaultNotesDir returns the default path for the notes directory.
func DefaultNotesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, AppDirName, NotesDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create notes dir: %v", err)
	}
	return dir, nil
}
