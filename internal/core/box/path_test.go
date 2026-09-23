package box

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "home", in: "~", want: home},
		{name: "under home", in: "~/notes/work", want: filepath.Join(home, "notes", "work")},
		{name: "relative", in: "notes", want: filepath.Join(cwd, "notes")},
		{name: "absolute is cleaned", in: "/a/b/../c/", want: "/a/c"},
		{name: "tilde not at start is kept", in: "/a/~b", want: "/a/~b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePath(tt.in)
			if err != nil {
				t.Fatalf("unexpected err occurred: %v", err)
			}
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
