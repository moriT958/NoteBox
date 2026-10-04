package preview

import (
	"strings"
	"sync"

	"notebox/internal/mermaid"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
)

// Renderer renders markdown for the preview. It's safe to call from
// concurrent commands; glamour's renderer itself is not.
type Renderer struct {
	mu sync.Mutex
	r  *glamour.TermRenderer
}

// NewRenderer returns a renderer for the given theme, which draws Mermaid
// diagrams in the given colors.
func NewRenderer(theme string, diagrams mermaid.Styles) (*Renderer, error) {
	style := styles.DarkStyle
	if theme == "light" {
		style = styles.LightStyle
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(0),
		glamour.WithCodeBlockRenderer(diagramRenderer(diagrams)),
	)
	if err != nil {
		return nil, err
	}
	return &Renderer{r: r}, nil
}

// diagramRenderer returns a code block renderer that draws Mermaid diagrams,
// leaving other code blocks, and diagrams it can't draw, as code.
func diagramRenderer(diagrams mermaid.Styles) func(language, code string) (string, bool) {
	return func(language, code string) (string, bool) {
		// Code blocks name their language case-insensitively, as the syntax
		// highlighter matches it.
		if strings.ToLower(language) != "mermaid" {
			return "", false
		}
		drawn, err := mermaid.Render(code, diagrams)
		if err != nil {
			return "", false
		}
		return drawn, true
	}
}

func (r *Renderer) Render(markdown string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Render(markdown)
}
