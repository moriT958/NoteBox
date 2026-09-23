package preview

import (
	"sync"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
)

// Renderer renders markdown for the preview. It's safe to call from
// concurrent commands; glamour's renderer itself is not.
type Renderer struct {
	mu sync.Mutex
	r  *glamour.TermRenderer
}

func NewRenderer(theme string) (*Renderer, error) {
	style := styles.DarkStyle
	if theme == "light" {
		style = styles.LightStyle
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return nil, err
	}
	return &Renderer{r: r}, nil
}

func (r *Renderer) Render(markdown string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Render(markdown)
}
