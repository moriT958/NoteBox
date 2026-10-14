package styles

import (
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeTheme(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write theme: %v", err)
	}
	return dir
}

func TestBuiltinPalettes(t *testing.T) {
	for name, p := range map[string]Palette{"dark": DarkPalette(), "light": LightPalette()} {
		for _, tk := range p.tokens() {
			if err := tk.check(*tk.value); err != nil {
				t.Errorf("%s: %s: %v", name, tk.name, err)
			}
		}
	}
}

func TestLoadTheme(t *testing.T) {
	t.Run("A built-in theme needs no file.", func(t *testing.T) {
		p, problems := LoadTheme("", "light")

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		if p != LightPalette() {
			t.Errorf("palette = %+v, want the light palette", p)
		}
	})

	t.Run("A file gives its values on top of its base.", func(t *testing.T) {
		dir := writeTheme(t, "mine", `{"base": "light", "accent": "#123456", "markdown": "dark"}`)

		p, problems := LoadTheme(dir, "mine")

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		want := LightPalette()
		want.Accent, want.Markdown = "#123456", "dark"
		if p != want {
			t.Errorf("palette = %+v, want %+v", p, want)
		}
	})

	t.Run("The base is dark when omitted.", func(t *testing.T) {
		dir := writeTheme(t, "mine", `{"accent": "#abc"}`)

		p, problems := LoadTheme(dir, "mine")

		if len(problems) != 0 {
			t.Fatalf("problems = %v, want none", problems)
		}
		want := DarkPalette()
		want.Accent = "#abc"
		if p != want {
			t.Errorf("palette = %+v, want %+v", p, want)
		}
	})

	t.Run("A file takes precedence over a built-in theme of the same name.", func(t *testing.T) {
		dir := writeTheme(t, "light", `{"base": "light", "accent": "#123456"}`)

		p, _ := LoadTheme(dir, "light")

		if p.Accent != "#123456" {
			t.Errorf("accent = %q, want the file's", p.Accent)
		}
	})

	t.Run("The name is not case sensitive.", func(t *testing.T) {
		p, problems := LoadTheme("", "Light")

		if len(problems) != 0 || p != LightPalette() {
			t.Errorf("got %+v, %v, want the light palette", p, problems)
		}
	})

	t.Run("An unknown theme falls back to dark.", func(t *testing.T) {
		p, problems := LoadTheme(t.TempDir(), "nope")

		if len(problems) != 1 {
			t.Fatalf("problems = %v, want one", problems)
		}
		if p != DarkPalette() {
			t.Errorf("palette = %+v, want the dark palette", p)
		}
	})

	t.Run("A name that could leave the directory is refused.", func(t *testing.T) {
		for _, name := range []string{"../x", "a/b", "", "a b"} {
			p, problems := LoadTheme(t.TempDir(), name)

			if len(problems) != 1 || p != DarkPalette() {
				t.Errorf("%q: got %+v, %v, want the dark palette and one problem", name, p, problems)
			}
		}
	})

	t.Run("A file that isn't JSON falls back to dark.", func(t *testing.T) {
		dir := writeTheme(t, "mine", `{"accent": `)

		p, problems := LoadTheme(dir, "mine")

		if len(problems) != 1 || p != DarkPalette() {
			t.Errorf("got %+v, %v, want the dark palette and one problem", p, problems)
		}
	})

	t.Run("An invalid base is reported and dark is used.", func(t *testing.T) {
		for _, content := range []string{`{"base": "mine"}`, `{"base": 1}`} {
			dir := writeTheme(t, "mine", content)

			p, problems := LoadTheme(dir, "mine")

			if len(problems) != 1 || p != DarkPalette() {
				t.Errorf("%s: got %+v, %v, want the dark palette and one problem", content, p, problems)
			}
		}
	})

	t.Run("Invalid values keep the base's and are reported one by one.", func(t *testing.T) {
		dir := writeTheme(t, "mine", `{
			"base": "light",
			"accent": "red",
			"cursor": 1,
			"markdown": "sepia",
			"muted": "#000000",
			"sparkle": "#ffffff"
		}`)

		p, problems := LoadTheme(dir, "mine")

		if len(problems) != 4 {
			t.Errorf("problems = %v, want four", problems)
		}
		want := LightPalette()
		want.Muted = "#000000"
		if p != want {
			t.Errorf("palette = %+v, want %+v", p, want)
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("The palette's colors reach the styles.", func(t *testing.T) {
		p := DarkPalette()
		p.Accent = "#112233"

		s := New(p)

		if got, want := s.FrameFocused.GetForeground(), color.Color(rgb(0x11, 0x22, 0x33)); got != want {
			t.Errorf("focused frame = %v, want %v", got, want)
		}
	})

	t.Run("A short color is the same as the long one.", func(t *testing.T) {
		short, long := DarkPalette(), DarkPalette()
		short.Accent, long.Accent = "#abc", "#aabbcc"

		if New(short).FrameFocused.GetForeground() != New(long).FrameFocused.GetForeground() {
			t.Errorf("#abc and #aabbcc differ")
		}
	})
}

func rgb(r, g, b uint8) color.Color { return color.RGBA{r, g, b, 0xff} }

func TestExportTheme(t *testing.T) {
	t.Run("An exported theme loads back as the built-in one.", func(t *testing.T) {
		dir := t.TempDir()
		for name, want := range map[string]Palette{"dark": DarkPalette(), "light": LightPalette()} {
			data, err := ExportTheme(name)
			if err != nil {
				t.Fatalf("ExportTheme(%q) error = %v", name, err)
			}
			if err := os.WriteFile(filepath.Join(dir, "mine-"+name+".json"), data, 0644); err != nil {
				t.Fatalf("failed to write theme: %v", err)
			}

			p, problems := LoadTheme(dir, "mine-"+name)

			if len(problems) != 0 {
				t.Errorf("%s: problems = %v, want none", name, problems)
			}
			if p != want {
				t.Errorf("%s: palette = %+v, want the built-in one", name, p)
			}
		}
	})

	t.Run("An unknown theme is an error.", func(t *testing.T) {
		if _, err := ExportTheme("nope"); err == nil {
			t.Errorf("ExportTheme() = nil, want an error")
		}
	})
}

func TestListThemes(t *testing.T) {
	t.Run("Built-in themes are listed without a directory.", func(t *testing.T) {
		got, err := ListThemes(filepath.Join(t.TempDir(), "missing"))

		want := []Theme{{Name: "dark"}, {Name: "light"}}
		if err != nil || !slices.Equal(got, want) {
			t.Errorf("got %v, %v, want %v", got, err, want)
		}
	})

	t.Run("Files are listed with the built-in themes, and win over them.", func(t *testing.T) {
		dir := writeTheme(t, "mine", `{}`)
		for _, name := range []string{"light.json", "Bad Name.json", "notes.txt"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(`{}`), 0644); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}
		}
		if err := os.Mkdir(filepath.Join(dir, "sub.json"), 0755); err != nil {
			t.Fatalf("failed to make dir: %v", err)
		}

		got, err := ListThemes(dir)

		want := []Theme{{Name: "dark"}, {Name: "light", File: true}, {Name: "mine", File: true}}
		if err != nil || !slices.Equal(got, want) {
			t.Errorf("got %v, %v, want %v", got, err, want)
		}
	})
}
