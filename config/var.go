package config

import (
	_ "embed"
)

const (
	CurrentVersion string = "v1.0"
	ConfigFile     string = "config.json"
)

var (
	Volume      string
	MetadataDir string
	Editor      string
	Grepcmd     string
)
