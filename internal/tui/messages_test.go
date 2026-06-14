package tui

import (
	"notebox/internal/note"
	"path/filepath"
	"testing"
)

func TestResolveBoxPath(t *testing.T) {
	const home = "/home/user"
	const cwd = "/current/dir"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tilde alone", "~", "/home/user"},
		{"tilde with subdir", "~/notes", "/home/user/notes"},
		{"dot", ".", "/current/dir"},
		{"dot slash subdir", "./notes", "/current/dir/notes"},
		{"dot dot", "..", "/current"},
		{"absolute", "/tmp/boxes", "/tmp/boxes"},
		{"relative no dot", "notes", "/current/dir/notes"},
		{"relative nested", "a/b/c", "/current/dir/a/b/c"},
		{"absolute with dot dot", "/tmp/../boxes", "/boxes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveBoxPath(tt.input, cwd, home)
			if got != tt.want {
				t.Errorf("resolveBoxPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewBoxFinalPath(t *testing.T) {
	const configDir = "/home/user/.notebox"

	tests := []struct {
		name         string
		title        string
		resolvedBase string
		want         string
	}{
		{"empty base uses configDir", "Work", "", filepath.Join(configDir, "Work")},
		{"with base appends title", "Work", "/tmp/boxes", "/tmp/boxes/Work"},
		{"nested base", "Tasks", "/home/user/projects", "/home/user/projects/Tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBoxFinalPath(tt.title, tt.resolvedBase, configDir)
			if got != tt.want {
				t.Errorf("newBoxFinalPath(%q, %q) = %q, want %q", tt.title, tt.resolvedBase, got, tt.want)
			}
		})
	}
}

func TestIsDuplicatePath(t *testing.T) {
	boxes := []note.Box{
		{ID: 1, Title: "A", Path: "/home/user/boxA"},
		{ID: 2, Title: "B", Path: "/home/user/boxB"},
	}

	tests := []struct {
		name  string
		path  string
		boxes []note.Box
		want  bool
	}{
		{"existing path", "/home/user/boxA", boxes, true},
		{"another existing path", "/home/user/boxB", boxes, true},
		{"new path", "/home/user/boxC", boxes, false},
		{"empty box list", "/home/user/boxA", nil, false},
		{"empty path no match", "", boxes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDuplicatePath(tt.path, tt.boxes)
			if got != tt.want {
				t.Errorf("isDuplicatePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
