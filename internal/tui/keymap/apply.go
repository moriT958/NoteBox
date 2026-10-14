package keymap

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
)

// Apply replaces the keys of the actions in overrides, keyed by action ID.
// Overrides that can't be used are skipped, keeping the default keys, and
// returned as problems.
func (k *KeyMap) Apply(overrides map[string][]string) []error {
	defaults := Default()
	byID := k.byID()
	defaultByID := defaults.byID()

	var problems []error
	overridden := map[string]bool{}
	for _, id := range slices.Sorted(maps.Keys(overrides)) {
		b, ok := byID[id]
		if !ok {
			problems = append(problems, fmt.Errorf("keymap: unknown action %q, ignored", id))
			continue
		}
		if err := validateKeys(overrides[id]); err != nil {
			problems = append(problems, fmt.Errorf("keymap: %q: %w, using the default keys", id, err))
			continue
		}
		setKeys(b, overrides[id])
		overridden[id] = true
	}

	// Going back to the defaults can make a new conflict with another
	// override, so look again until none is left. The defaults never
	// conflict, so every round puts back at least one override.
	for {
		a, b, shared, ok := k.conflict(overridden)
		if !ok {
			break
		}
		problems = append(problems, fmt.Errorf("keymap: %q and %q both use %q, using their default keys", a, b, shared))
		for _, id := range []string{a, b} {
			if overridden[id] {
				*byID[id] = *defaultByID[id]
				delete(overridden, id)
			}
		}
	}

	return problems
}

func (k *KeyMap) byID() map[string]*key.Binding {
	m := map[string]*key.Binding{}
	for _, a := range k.actions() {
		m[a.id] = a.b
	}
	return m
}

// groups lists the actions that are active at the same time, so they must
// not share a key. An ID ending in ".*" stands for every action under it.
var groups = [][]string{
	{"global.*", "list.*"},
	{"global.*", "preview.*"},
	{"global.quit", "rename.*"},
	{"global.quit", "global.boxes", "dialog.*", "boxes.*"},
	{"global.quit", "global.boxes", "dialog.confirm", "dialog.cancel", "form.*"},
	{"global.quit", "global.boxes", "dialog.confirm", "confirm.no"},
}

// conflict returns two actions active at the same time that share a key,
// at least one of them overridden.
func (k *KeyMap) conflict(overridden map[string]bool) (a, b, shared string, ok bool) {
	actions := k.actions()
	for _, group := range groups {
		users := map[string]string{}
		for _, act := range actions {
			if !inGroup(act.id, group) {
				continue
			}
			for _, s := range act.b.Keys() {
				other, used := users[s]
				if used && other != act.id && (overridden[other] || overridden[act.id]) {
					return other, act.id, s, true
				}
				if !used {
					users[s] = act.id
				}
			}
		}
	}
	return "", "", "", false
}

func inGroup(id string, group []string) bool {
	for _, g := range group {
		if g == id {
			return true
		}
		if prefix, ok := strings.CutSuffix(g, "*"); ok && strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

// modifiers are in the order key presses print them in.
var modifiers = []string{"ctrl", "alt", "shift", "meta", "hyper", "super"}

// namedKeys are the keys written by name rather than by their character.
var namedKeys = func() map[string]bool {
	m := map[string]bool{}
	for _, name := range []string{
		"enter", "tab", "backspace", "esc", "space",
		"up", "down", "left", "right",
		"home", "end", "pgup", "pgdown", "insert", "delete",
	} {
		m[name] = true
	}
	for i := 1; i <= 20; i++ {
		m[fmt.Sprintf("f%d", i)] = true
	}
	return m
}()

// validateKeys checks that keys are written the way key presses print, so
// they can match: modifiers in a fixed order, then a key name or a single
// character.
func validateKeys(keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("no keys")
	}
	for _, s := range keys {
		if err := validateKey(s); err != nil {
			return fmt.Errorf("invalid key %q: %w", s, err)
		}
	}
	return nil
}

func validateKey(s string) error {
	if s == "" {
		return fmt.Errorf("empty")
	}
	// A lone "+" is the plus key, not a separator.
	name := s
	var mods []string
	if s != "+" {
		parts := strings.Split(s, "+")
		if strings.HasSuffix(s, "++") {
			parts = append(parts[:len(parts)-2], "+")
		}
		name, mods = parts[len(parts)-1], parts[:len(parts)-1]
	}

	last := -1
	for _, mod := range mods {
		i := slices.Index(modifiers, mod)
		if i < 0 {
			return fmt.Errorf("unknown modifier %q", mod)
		}
		if i <= last {
			return fmt.Errorf("write modifiers in the order %s", strings.Join(modifiers, "+"))
		}
		last = i
	}

	if namedKeys[name] {
		return nil
	}
	r, size := utf8.DecodeRuneInString(name)
	if name == "" || size != len(name) || !unicode.IsPrint(r) || unicode.IsSpace(r) {
		return fmt.Errorf("unknown key %q", name)
	}
	// A shifted character prints as itself: "A", not "shift+a".
	if slices.Contains(mods, "shift") {
		return fmt.Errorf("write the shifted character instead of shift+%s", name)
	}
	return nil
}

// helpKey is how keys show in the help: the first two, arrows as symbols.
func helpKey(keys []string) string {
	shown := keys[:min(len(keys), 2)]
	labels := make([]string, len(shown))
	for i, s := range shown {
		mods, name := "", s
		if j := strings.LastIndex(s, "+"); j > 0 && j < len(s)-1 {
			mods, name = s[:j+1], s[j+1:]
		}
		if arrow, ok := arrows[name]; ok {
			name = arrow
		}
		labels[i] = mods + name
	}
	return strings.Join(labels, "/")
}

var arrows = map[string]string{"up": "↑", "down": "↓", "left": "←", "right": "→"}
