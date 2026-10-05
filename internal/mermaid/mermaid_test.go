package mermaid

import (
	"errors"
	"notebox/internal/mermaid/internal/diagram"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// withDiagram registers a diagram type for the duration of a test.
func withDiagram(t *testing.T, keyword string, draw drawFunc) {
	t.Helper()
	diagrams[keyword] = draw
	t.Cleanup(func() { delete(diagrams, keyword) })
}

// drawText returns a drawFunc that draws s as text, whatever the source.
func drawText(s string) drawFunc {
	return func(string) (*diagram.Canvas, error) {
		c := diagram.New(ansi.StringWidth(s), 1)
		c.Text(0, 0, s, diagram.RoleText)
		return c, nil
	}
}

func TestRenderDispatch(t *testing.T) {
	withDiagram(t, "test", drawText("drawn"))

	tests := []struct {
		name string
		src  string
	}{
		{"declaration", "test\nA"},
		{"declaration with arguments", "test LR;\nA"},
		{"declaration followed by a semicolon", "test;A"},
		{"tab after the keyword", "test\tLR"},
		{"leading blank lines and comments", "\n  \n%% comment\n%%{init: {}}%%\n  test\n"},
		{"front matter", "---\ntitle: x\nconfig: {}\n---\ntest\n"},
		{"front matter after blank lines", "\n---\ntitle: x\n---\n\ntest\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.src, Styles{})
			if err != nil || got != "drawn" {
				t.Errorf("Render() = %q, %v; want %q, nil", got, err, "drawn")
			}
		})
	}
}

func TestRenderUnsupported(t *testing.T) {
	withDiagram(t, "test", drawText("drawn"))

	for _, src := range []string{
		"",
		"%% only a comment",
		"pie\n  \"a\": 1",
		"Test",
		"tests",
		"---\ntest\n",
	} {
		if _, err := Render(src, Styles{}); !errors.Is(err, ErrUnsupported) {
			t.Errorf("Render(%q) error = %v, want ErrUnsupported", src, err)
		}
	}
}

func TestRenderPanic(t *testing.T) {
	withDiagram(t, "test", func(string) (*diagram.Canvas, error) { panic("boom") })

	got, err := Render("test", Styles{})
	if err == nil || got != "" {
		t.Errorf("Render() = %q, %v; want an error", got, err)
	}
}

func TestRenderError(t *testing.T) {
	withDiagram(t, "test", func(string) (*diagram.Canvas, error) { return nil, ErrSyntax })

	if got, err := Render("test", Styles{}); !errors.Is(err, ErrSyntax) || got != "" {
		t.Errorf("Render() = %q, %v; want ErrSyntax", got, err)
	}
}

func TestRenderStyles(t *testing.T) {
	withDiagram(t, "test", drawText("drawn"))

	got, err := Render("test", Styles{Text: ansi.IndexedColor(1)})
	if want := "\x1b[38;5;1mdrawn\x1b[39m"; err != nil || got != want {
		t.Errorf("Render() = %q, %v; want %q, nil", got, err, want)
	}
}
