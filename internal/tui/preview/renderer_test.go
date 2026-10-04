package preview

import (
	"strings"
	"testing"

	"notebox/internal/mermaid"

	"github.com/charmbracelet/x/ansi"
)

func newTestRenderer(t *testing.T, diagrams mermaid.Styles) *Renderer {
	t.Helper()
	r, err := NewRenderer("dark", diagrams)
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}
	return r
}

func TestRendererDiagrams(t *testing.T) {
	r := newTestRenderer(t, mermaid.Styles{})
	drawn, err := mermaid.Render("graph TD\nA --> B", mermaid.Styles{})
	if err != nil {
		t.Fatalf("mermaid.Render() error = %v", err)
	}

	tests := []struct {
		name string
		md   string
		// want is in the output when the diagram is drawn, and source when
		// it is left as code.
		want, source string
	}{
		{name: "backtick fence", md: "```mermaid\ngraph TD\nA --> B\n```\n", want: drawn},
		{name: "tilde fence", md: "~~~mermaid\ngraph TD\nA --> B\n~~~\n", want: drawn},
		{name: "language in any case", md: "```Mermaid\ngraph TD\nA --> B\n```\n", want: drawn},
		{name: "unsupported diagram", md: "```mermaid\npie\n  \"a\": 1\n```\n", source: `"a": 1`},
		{name: "unsupported syntax", md: "```mermaid\ngraph TD\nsubgraph one\nA --> B\nend\n```\n", source: "subgraph one"},
		{name: "other languages", md: "```go\nA --> B\n```\n", source: "A --> B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := r.Render(tt.md)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			plain := ansi.Strip(out)
			if tt.want != "" {
				for line := range strings.SplitSeq(tt.want, "\n") {
					if !strings.Contains(plain, line) {
						t.Errorf("diagram line %q missing from\n%s", line, plain)
					}
				}
			}
			if tt.source != "" && !strings.Contains(plain, tt.source) {
				t.Errorf("source %q missing from\n%s", tt.source, plain)
			}
		})
	}
}

func TestRendererDiagramColors(t *testing.T) {
	r := newTestRenderer(t, mermaid.Styles{
		Border:    ansi.IndexedColor(1),
		Text:      ansi.IndexedColor(2),
		Edge:      ansi.IndexedColor(3),
		EdgeLabel: ansi.IndexedColor(4),
	})
	out, err := r.Render("```mermaid\ngraph TD\nA -->|yes| B\n```\n")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, seq := range []string{"\x1b[38;5;1m┌", "\x1b[38;5;2mA", "\x1b[38;5;3m▼", "\x1b[38;5;4myes"} {
		if !strings.Contains(out, seq) {
			t.Errorf("%q missing from %q", seq, out)
		}
	}
}
