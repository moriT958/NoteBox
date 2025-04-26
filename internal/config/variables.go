package config

import (
	"path/filepath"
	// "github.com/adrg/xdg"
)

// var (
// 	userHome          = xdg.Home
// 	noteBoxHome       = filepath.Join(xdg.ConfigHome, "notebox")
// 	configFile        = filepath.Join(noteBoxHome, "config.json")
// 	defaultFileVolume = filepath.Join(xdg.DataHome, "notebox")
// 	noteBoxStateDir   = filepath.Join(xdg.StateHome, "notebox")
// 	LogFile           = filepath.Join(noteBoxStateDir, "notebox.log")
// )

/* For Debugging */
var (
	userHome          = "."
	noteBoxHome       = filepath.Join(".", ".config", "notebox")
	configFile        = filepath.Join(noteBoxHome, "config.json")
	defaultFileVolume = filepath.Join(".", ".volume", "notebox")
	noteBoxStateDir   = filepath.Join(".", ".state", "notebox")
	LogFile           = filepath.Join(noteBoxStateDir, "notebox.log")
)
