package utils

import (
	"fmt"
	"os"
)

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get home dir: %v", err)
		os.Exit(1)
	}
	return home
}
