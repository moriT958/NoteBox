package mermaid

import (
	"errors"
	"testing"
)

// withDiagram registers a diagram type for the duration of a test.
func withDiagram(t *testing.T, keyword string, draw drawFunc) {
	t.Helper()
	diagrams[keyword] = draw
	t.Cleanup(func() { delete(diagrams, keyword) })
}

func TestRenderDispatch(t *testing.T) {
	withDiagram(t, "test", func(src string) (string, error) { return "drawn", nil })

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
			got, err := Render(tt.src)
			if err != nil || got != "drawn" {
				t.Errorf("Render() = %q, %v; want %q, nil", got, err, "drawn")
			}
		})
	}
}

func TestRenderUnsupported(t *testing.T) {
	withDiagram(t, "test", func(src string) (string, error) { return "drawn", nil })

	for _, src := range []string{
		"",
		"%% only a comment",
		"pie\n  \"a\": 1",
		"Test",
		"tests",
		"---\ntest\n",
	} {
		if _, err := Render(src); !errors.Is(err, ErrUnsupported) {
			t.Errorf("Render(%q) error = %v, want ErrUnsupported", src, err)
		}
	}
}

func TestRenderPanic(t *testing.T) {
	withDiagram(t, "test", func(src string) (string, error) { panic("boom") })

	got, err := Render("test")
	if err == nil || got != "" {
		t.Errorf("Render() = %q, %v; want an error", got, err)
	}
}
