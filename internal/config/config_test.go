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
	t.Run("A missing file gives the defaults and isn't created.", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ConfigFileName)

		cfg, problems := Load(path)

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		if !reflect.DeepEqual(cfg, defaultConfig()) {
			t.Errorf("cfg = %+v, want the defaults", *cfg)
		}
		if _, err := os.Stat(path); err == nil {
			t.Errorf("Load created the file")
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

func TestWriteDefault(t *testing.T) {
	t.Run("The file has the defaults and loads back.", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), ConfigFileName)

		if err := WriteDefault(path); err != nil {
			t.Fatalf("WriteDefault() error = %v", err)
		}

		cfg, problems := Load(path)
		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		if !reflect.DeepEqual(cfg, defaultConfig()) {
			t.Errorf("cfg = %+v, want the defaults", *cfg)
		}
		data, _ := os.ReadFile(path)
		if want := "{\n  \"editor\": \"vim\",\n  \"theme\": \"dark\"\n}\n"; string(data) != want {
			t.Errorf("file =\n%s\nwant\n%s", data, want)
		}
	})

	t.Run("An existing file is left alone.", func(t *testing.T) {
		path := writeConfig(t, `{"editor": "nvim"}`)

		if err := WriteDefault(path); err != nil {
			t.Fatalf("WriteDefault() error = %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) != `{"editor": "nvim"}` {
			t.Errorf("file = %s, want it unchanged", data)
		}
	})

	t.Run("A directory that doesn't exist is an error.", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", ConfigFileName)

		if err := WriteDefault(path); err == nil {
			t.Errorf("WriteDefault() = nil, want an error")
		}
	})
}
