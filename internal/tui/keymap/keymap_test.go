package keymap

import (
	"slices"
	"testing"
)

// allOverridden marks every action as overridden, so conflict reports any
// shared key.
func allOverridden(k *KeyMap) map[string]bool {
	m := map[string]bool{}
	for _, a := range k.actions() {
		m[a.id] = true
	}
	return m
}

func TestDefault(t *testing.T) {
	t.Run("Every action has its own ID.", func(t *testing.T) {
		k := Default()
		seen := map[string]bool{}
		for _, a := range k.actions() {
			if seen[a.id] {
				t.Errorf("duplicate ID %q", a.id)
			}
			seen[a.id] = true
		}
	})

	t.Run("The defaults don't conflict.", func(t *testing.T) {
		k := Default()
		if a, b, s, ok := k.conflict(allOverridden(&k)); ok {
			t.Errorf("%q and %q both use %q", a, b, s)
		}
	})
}

func TestApply(t *testing.T) {
	t.Run("An override replaces the keys and the help.", func(t *testing.T) {
		k := Default()

		problems := k.Apply(map[string][]string{"list.new": {"a", "N"}})

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		if got := k.List.New.Keys(); !slices.Equal(got, []string{"a", "N"}) {
			t.Errorf("keys = %v, want [a N]", got)
		}
		if got := k.List.New.Help(); got.Key != "a/N" || got.Desc != "new" {
			t.Errorf("help = %+v, want a/N new", got)
		}
	})

	t.Run("An unknown action is reported and ignored.", func(t *testing.T) {
		k := Default()

		problems := k.Apply(map[string][]string{"list.nwe": {"a"}})

		if len(problems) != 1 {
			t.Fatalf("problems = %v, want one", problems)
		}
		if got := k.List.New.Keys(); !slices.Equal(got, []string{"n"}) {
			t.Errorf("keys = %v, want the default", got)
		}
	})

	t.Run("Invalid keys keep the default.", func(t *testing.T) {
		for _, keys := range [][]string{nil, {}, {"n", "shift+ctrl+n"}} {
			k := Default()

			problems := k.Apply(map[string][]string{"list.new": keys})

			if len(problems) != 1 {
				t.Errorf("%v: problems = %v, want one", keys, problems)
			}
			if got := k.List.New.Keys(); !slices.Equal(got, []string{"n"}) {
				t.Errorf("%v: keys = %v, want the default", keys, got)
			}
		}
	})

	t.Run("Actions sharing a key go back to their defaults.", func(t *testing.T) {
		k := Default()

		problems := k.Apply(map[string][]string{"list.new": {"d"}})

		if len(problems) != 1 {
			t.Fatalf("problems = %v, want one", problems)
		}
		if got := k.List.New.Keys(); !slices.Equal(got, []string{"n"}) {
			t.Errorf("keys = %v, want the default", got)
		}
	})

	t.Run("A key shared with a global action conflicts.", func(t *testing.T) {
		k := Default()

		problems := k.Apply(map[string][]string{"preview.edit": {"ctrl+b"}})

		if len(problems) != 1 {
			t.Fatalf("problems = %v, want one", problems)
		}
	})

	t.Run("Actions never active together may share a key.", func(t *testing.T) {
		k := Default()

		problems := k.Apply(map[string][]string{"list.new": {"w"}})

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
	})

	t.Run("A conflict made by going back to a default is resolved too.", func(t *testing.T) {
		k := Default()

		// delete and edit share x; delete going back to d then conflicts
		// with new.
		problems := k.Apply(map[string][]string{
			"list.new":    {"d"},
			"list.delete": {"x"},
			"list.edit":   {"x"},
		})

		if len(problems) != 2 {
			t.Fatalf("problems = %v, want two", problems)
		}
		if a, b, s, ok := k.conflict(allOverridden(&k)); ok {
			t.Errorf("%q and %q still both use %q", a, b, s)
		}
	})
}

func TestValidateKey(t *testing.T) {
	valid := []string{"a", "A", "?", "/", "+", "ctrl++", "ctrl+c", "ctrl+alt+x", "shift+tab", "pgdown", "f5", "space", "é"}
	for _, s := range valid {
		if err := validateKey(s); err != nil {
			t.Errorf("validateKey(%q) = %v, want nil", s, err)
		}
	}
	invalid := []string{"", " ", "ab", "ctl+c", "alt+ctrl+x", "shift+a", "ctrl+", "Enter"}
	for _, s := range invalid {
		if err := validateKey(s); err == nil {
			t.Errorf("validateKey(%q) = nil, want an error", s)
		}
	}
}

func TestHelpKey(t *testing.T) {
	tests := []struct {
		keys []string
		want string
	}{
		{[]string{"up", "k", "ctrl+p"}, "↑/k"},
		{[]string{"shift+left"}, "shift+←"},
		{[]string{"pgup", "b"}, "pgup/b"},
		{[]string{"ctrl++"}, "ctrl++"},
	}
	for _, tt := range tests {
		if got := helpKey(tt.keys); got != tt.want {
			t.Errorf("helpKey(%v) = %q, want %q", tt.keys, got, tt.want)
		}
	}
}
