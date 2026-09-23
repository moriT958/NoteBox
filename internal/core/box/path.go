package box

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// normalizePath expands a leading ~ and makes p absolute and clean, so that
// every stored box path means the same directory regardless of the working
// directory, and can be compared as-is to find duplicates.
func normalizePath(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand ~: %w", err)
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return filepath.Abs(p)
}
