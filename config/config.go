package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	cfg  *Config
	once sync.Once
)

type Config struct {
	HomeDir     string `json:"-"`
	CfgDir      string `json:"-"`
	MetaDataDir string `json:"-"`
	Volume      string `json:"volume"`
	Editor      string `json:"editor"`
	Grepcmd     string `json:"grepcmd"`
}

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func GetConfig() *Config {
	once.Do(func() {
		cfg = loadDefaultConfig()
	})

	return cfg
}

func loadDefaultConfig() *Config {
	cfg = new(Config)
	cfg.HomeDir = getHomeDir()
	cfg.CfgDir = filepath.Join(getHomeDir(), ".config", "notebox", "config.json")
	cfg.MetaDataDir = filepath.Join(getHomeDir(), ".config", "notebox", ".metadata.sqlite")

	if err := os.MkdirAll(filepath.Join(cfg.HomeDir, ".config", "notebox", "files"), 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := os.Stat(cfg.CfgDir); os.IsNotExist(err) {
		cfg.Volume = filepath.Join(getHomeDir(), ".config", "notebox", "files")
		cfg.Editor = "vim"
		cfg.Grepcmd = "grep"

		fp, err := os.Create(cfg.CfgDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := json.NewEncoder(fp).Encode(cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer fp.Close()

	}

	fp, err := os.Open(cfg.CfgDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer fp.Close()

	if err := json.NewDecoder(fp).Decode(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return cfg
}

func CfgDir() string {
	return cfg.CfgDir
}

func Volume() string {
	return cfg.Volume
}

func Editor() string {
	return cfg.Editor
}

func MetaDataDir() string {
	return cfg.MetaDataDir
}
