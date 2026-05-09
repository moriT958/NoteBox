package note

import (
	"os"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
)

type NoteRenderer interface {
	RenderNote(Note) (string, error)
}

type GlamourRenderer struct {
	*glamour.TermRenderer
}

func NewGlamourRenderer(theme string) (GlamourRenderer, error) {
	glamourTheme := styles.DarkStyle
	if theme == "light" {
		glamourTheme = styles.LightStyle
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(glamourTheme),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return GlamourRenderer{}, err
	}
	return GlamourRenderer{TermRenderer: r}, nil
}

func (g GlamourRenderer) RenderNote(n Note) (string, error) {
	if n == (Note{}) {
		return "(( No Content ))", nil
	}

	b, err := os.ReadFile(n.Path)
	if err != nil {
		return "", err
	}

	return g.TermRenderer.Render(string(b))
}
