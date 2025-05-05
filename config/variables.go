package config

// var (
// 	userHome          = xdg.Home
// 	noteBoxHome       = filepath.Join(xdg.ConfigHome, "notebox")
// 	configFile        = filepath.Join(noteBoxHome, "config.json")
// 	defaultFileVolume = filepath.Join(xdg.DataHome, "notebox")
// 	noteBoxStateDir   = filepath.Join(xdg.StateHome, "notebox")
// 	LogFile           = filepath.Join(noteBoxStateDir, "notebox.log")
// )

const (
	CurrentVersion string = "version 0.0"
	BaseDir        string = "./notes"
	DummyNotePath  string = "./dummy.md"
	DefaultEditor  string = "nvim"
	LogfilePath    string = "notebox.log"
)
