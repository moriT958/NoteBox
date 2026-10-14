package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	// Set by goreleaser
	CurrentVersion string = "dev"
)

type Config struct {
	Editor string `json:"editor"`
	Theme  string `json:"theme"`
}

// Load reads the config file at path, creating it with the defaults when
// it doesn't exist. It always returns a usable config: when the file can't
// be read, the defaults are used and the reasons are returned.
func Load(path string) (*Config, []error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := writeDefault(path); err != nil {
			return cfg, []error{err}
		}
		return cfg, nil
	}
	if err != nil {
		return cfg, []error{fmt.Errorf("failed to read config file: %w", err)}
	}

	// Fields missing from the file keep their defaults.
	if err := json.Unmarshal(data, cfg); err != nil {
		return defaultConfig(), []error{fmt.Errorf("%s: invalid JSON, using the default config: %w", path, err)}
	}
	cfg.Theme = strings.ToLower(cfg.Theme)

	return cfg, nil
}

func writeDefault(path string) error {
	data, err := json.MarshalIndent(defaultConfig(), "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	return nil
}

func defaultConfig() *Config {
	return &Config{
		Editor: DefaultEditor,
		Theme:  defaultTheme,
	}
}

// LoadLastBoxID reads the last used box ID from the state file.
// Returns "" (no preference) when the file does not exist.
func LoadLastBoxID() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(home, AppDirName, StateFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveLastBoxID writes the given box ID to the state file.
func SaveLastBoxID(id string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, AppDirName, StateFileName), []byte(id), 0644)
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
