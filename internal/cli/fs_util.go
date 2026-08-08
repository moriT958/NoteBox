package cli

import (
	"os"
	"strings"
)

// shortenHomePath replaces the home directory prefix with "~".
func shortenHomePath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + strings.TrimPrefix(path, home)
}
