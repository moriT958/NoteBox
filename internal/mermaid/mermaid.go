// Package mermaid draws Mermaid diagrams as text with box-drawing characters.
package mermaid

import (
	"errors"
	"fmt"

	"notebox/internal/mermaid/diagram"
	"notebox/internal/mermaid/flowchart"
)

// Styles are the colors of the parts of a diagram. With the zero value, a
// diagram is drawn without colors. A part without a color is drawn in the
// terminal's default color rather than in the color of the text around the
// diagram, so a colored diagram should give every part a color.
type Styles = diagram.Styles

var (
	// ErrUnsupported is returned for diagram types, and syntax within them,
	// that can't be drawn yet.
	ErrUnsupported = diagram.ErrUnsupported
	// ErrSyntax is returned for diagrams that can't be parsed.
	ErrSyntax = diagram.ErrSyntax
	// ErrTooLarge is returned for diagrams too large to draw.
	ErrTooLarge = diagram.ErrTooLarge
)

// maxTextSize is Mermaid's own limit on the length of a diagram's source by
// default.
const maxTextSize = 50000

// errPanic reports a bug: a panic while drawing a diagram.
var errPanic = errors.New("mermaid: panic")

// A drawFunc draws a diagram of one type from its whole source.
type drawFunc func(src string) (*diagram.Canvas, error)

// diagrams maps the keyword that opens a diagram to the function drawing it.
var diagrams = map[string]drawFunc{
	"graph":     flowchart.Draw,
	"flowchart": flowchart.Draw,
}

// Render draws a Mermaid diagram in the given colors. Diagrams that can't be
// drawn are reported with one of the errors above, so that the caller can
// show the source instead.
func Render(src string, styles Styles) (out string, err error) {
	defer func() {
		if r := recover(); r != nil {
			out, err = "", fmt.Errorf("%w: %v", errPanic, r)
		}
	}()

	if len(src) > maxTextSize {
		return "", fmt.Errorf("%w: longer than %d bytes", ErrTooLarge, maxTextSize)
	}
	keyword := diagramKeyword(src)
	draw, ok := diagrams[keyword]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnsupported, keyword)
	}
	c, err := draw(src)
	if err != nil {
		return "", err
	}
	return c.Render(styles), nil
}
