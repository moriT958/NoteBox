package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Volume string `json:"volume"`
	Editor string `json:"editor"`
}

func LoadConfig() (*Config, error) {
	cfg := new(Config)

	// make note box home directory
	if err := os.MkdirAll(noteBoxHome, 0755); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(defaultFileVolume, 0755); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(noteBoxStateDir, 0755); err != nil {
		return nil, err
	}

	// Check if Config file already exits. If not exit then create config file.
	if _, err := os.Stat(configFile); err != nil {
		cfg = loadDefaultConfig()

		fp, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		defer fp.Close()

		if err := json.NewEncoder(fp).Encode(cfg); err != nil {
			return nil, err
		}
	}

	// if config file already exits.
	fp, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	// decode config file to struct
	if err := json.NewDecoder(fp).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadDefaultConfig() *Config {
	cfg := &Config{
		Volume: defaultFileVolume,
		Editor: "nvim",
	}
	return cfg
}
