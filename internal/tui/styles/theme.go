package styles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// themeName matches the names of themes: lowercase letters and digits,
// separated by single hyphens or underscores.
var themeName = regexp.MustCompile(`^[a-z0-9]+([-_][a-z0-9]+)*$`)

// LoadTheme returns the palette of the theme called name: the file
// <dir>/<name>.json if there is one, otherwise the built-in theme of that
// name. A file takes precedence over a built-in theme of the same name. An
// empty dir means no theme files.
//
// It always returns a usable palette. When the theme can't be used, that is
// the dark theme; when only some values of a theme file can't be used, those
// take the value of its base. What went wrong is returned as problems.
func LoadTheme(dir, name string) (Palette, []error) {
	name = strings.ToLower(name)
	if !themeName.MatchString(name) {
		return DarkPalette(), []error{fmt.Errorf("theme %q: use lowercase letters, numbers, hyphens, and underscores only, using the dark theme", name)}
	}

	if dir != "" {
		path := filepath.Join(dir, name+".json")
		data, err := os.ReadFile(path)
		switch {
		case err == nil:
			return parseTheme(filepath.Base(path), data)
		case !errors.Is(err, fs.ErrNotExist):
			return DarkPalette(), []error{fmt.Errorf("theme %q: %w, using the dark theme", name, err)}
		}
	}

	if p, ok := builtin(name); ok {
		return p, nil
	}
	return DarkPalette(), []error{fmt.Errorf("theme %q not found, using the dark theme", name)}
}

// parseTheme reads a theme file: the values it gives on top of its base.
func parseTheme(file string, data []byte) (Palette, []error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return DarkPalette(), []error{fmt.Errorf("theme file %s: %w, using the dark theme", file, err)}
	}

	var problems []error
	p := DarkPalette()
	if v, ok := raw["base"]; ok {
		var base string
		if err := json.Unmarshal(v, &base); err != nil {
			base = ""
		}
		if b, ok := builtin(base); ok {
			p = b
		} else {
			problems = append(problems, fmt.Errorf(`theme file %s: invalid base %s, use "dark" or "light", using "dark"`, file, v))
		}
	}

	known := map[string]bool{"base": true}
	for _, t := range p.tokens() {
		known[t.name] = true
		v, ok := raw[t.name]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			problems = append(problems, fmt.Errorf("theme file %s: %q must be a string, using the base's value", file, t.name))
			continue
		}
		if err := t.check(s); err != nil {
			problems = append(problems, fmt.Errorf("theme file %s: %q: %w, using the base's value", file, t.name, err))
			continue
		}
		*t.value = s
	}

	for _, key := range slices.Sorted(maps.Keys(raw)) {
		if !known[key] {
			problems = append(problems, fmt.Errorf("theme file %s: unknown value %q, ignored", file, key))
		}
	}

	return p, problems
}
