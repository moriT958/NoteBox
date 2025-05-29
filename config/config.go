package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"notebox/utils"
	"os"
	"path/filepath"
	"sync"
)

var (
	CurrentVersion string = "version 0.0"
)

var (
	cfg  *Config
	once sync.Once
)

type Config struct {
	Editor       string `json:"editor"`
	NotesDir     string `json:"notesdir"`
	DummyNoteDir string `json:"-"`
}

func Load() error {
	var err error
	filename := filepath.Join(utils.HomeDir(), ".notebox", "config.json")

	if _, err := os.Stat(filename); errors.Is(err, fs.ErrNotExist) {
		fp, createErr := os.Create(filename)
		if createErr != nil {
			return fmt.Errorf("failed to create config file: %v", createErr)
		}
		defer fp.Close()

		enc := json.NewEncoder(fp)
		enc.SetIndent("", "  ")
		if err := enc.Encode(defualtConfig()); err != nil {
			return fmt.Errorf("failed to encode config file: %v", err)
		}
	}

	once.Do(func() {
		cfg = new(Config)

		file, openErr := os.Open(filename)
		if openErr != nil {
			err = openErr
			return
		}
		defer file.Close()

		if decodeErr := json.NewDecoder(file).Decode(&cfg); decodeErr != nil {
			err = decodeErr
			return
		}

		cfg.DummyNoteDir = defualtConfig().DummyNoteDir
	})

	return err
}

func defualtConfig() *Config {
	notesdir := filepath.Join(utils.HomeDir(), ".notebox", "notes")
	dummyNoteDir := filepath.Join(utils.HomeDir(), ".notebox", "dummy.md")

	if err := os.MkdirAll(notesdir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create notes dir:", err)
		os.Exit(1)
	}

	fp, err := os.Create(dummyNoteDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create dummy note:", err)
		os.Exit(1)
	}
	defer fp.Close()
	fmt.Fprint(fp, "(( No Note Selected ))")

	return &Config{
		Editor:       "vim",
		NotesDir:     notesdir,
		DummyNoteDir: dummyNoteDir,
	}
}

func GetConfig() (*Config, error) {
	if cfg == nil {
		return nil, errors.New("config not initialized")
	}
	return cfg, nil
}

// TODO:
// this accessor methods will be removed.
func Editor() string       { return cfg.Editor }
func NotesDir() string     { return cfg.NotesDir }
func DummyNoteDir() string { return cfg.DummyNoteDir }
