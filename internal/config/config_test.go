package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ConfigFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	t.Run("A missing file is created with the defaults.", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ConfigFileName)

		cfg, problems := Load(path)

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		if !reflect.DeepEqual(cfg, defaultConfig()) {
			t.Errorf("cfg = %+v, want the defaults", *cfg)
		}
		created, problems := Load(path)
		if len(problems) != 0 || !reflect.DeepEqual(created, defaultConfig()) {
			t.Errorf("created file loads as %+v, %v, want the defaults", *created, problems)
		}
	})

	t.Run("Fields missing from the file keep their defaults.", func(t *testing.T) {
		path := writeConfig(t, `{"theme": "Light"}`)

		cfg, problems := Load(path)

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		want := &Config{Editor: DefaultEditor, Theme: "light"}
		if !reflect.DeepEqual(cfg, want) {
			t.Errorf("cfg = %+v, want %+v", *cfg, want)
		}
	})

	t.Run("The keymap is read as given.", func(t *testing.T) {
		path := writeConfig(t, `{"keymap": {"list.new": ["n", "a"]}}`)

		cfg, problems := Load(path)

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		want := map[string][]string{"list.new": {"n", "a"}}
		if !reflect.DeepEqual(cfg.Keymap, want) {
			t.Errorf("keymap = %v, want %v", cfg.Keymap, want)
		}
	})

	t.Run("Invalid JSON falls back to the defaults and is reported.", func(t *testing.T) {
		path := writeConfig(t, `{"editor": "nvim",`)

		cfg, problems := Load(path)

		if len(problems) != 1 {
			t.Fatalf("problems = %v, want one", problems)
		}
		if !reflect.DeepEqual(cfg, defaultConfig()) {
			t.Errorf("cfg = %+v, want the defaults", *cfg)
		}
	})
}
