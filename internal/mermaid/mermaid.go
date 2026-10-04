// Package mermaid draws Mermaid diagrams as text with box-drawing characters.
package mermaid

import (
	"errors"
	"fmt"
	"image/color"
	"strings"
)

var (
	// ErrUnsupported is returned for diagram types, and syntax within them,
	// that can't be drawn yet.
	ErrUnsupported = errors.New("mermaid: unsupported diagram")
	// ErrSyntax is returned for diagrams that can't be parsed.
	ErrSyntax = errors.New("mermaid: syntax error")
	// ErrTooLarge is returned for diagrams too large to draw.
	ErrTooLarge = errors.New("mermaid: diagram too large")
)

// Styles are the colors of the parts of a diagram. With the zero value, a
// diagram is drawn without colors. A part without a color is drawn in the
// terminal's default color rather than in the color of the text around the
// diagram, so a colored diagram should give every part a color.
type Styles struct {
	// Border is the color of the outlines of nodes and groups.
	Border color.Color
	// Text is the color of the labels of nodes and groups.
	Text color.Color
	// Edge is the color of the lines of edges and their heads.
	Edge color.Color
	// EdgeLabel is the color of the labels of edges.
	EdgeLabel color.Color
}

// A drawFunc draws a diagram of one type from its whole source.
type drawFunc func(src string) (*canvas, error)

// diagrams maps the keyword that opens a diagram to the function drawing it.
var diagrams = map[string]drawFunc{}

// Render draws a Mermaid diagram in the given colors. Diagrams that can't be
// drawn are reported with one of the errors above, so that the caller can
// show the source instead.
func Render(src string, styles Styles) (out string, err error) {
	defer func() {
		if r := recover(); r != nil {
			out, err = "", fmt.Errorf("mermaid: panic: %v", r)
		}
	}()

	keyword := diagramKeyword(src)
	draw, ok := diagrams[keyword]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnsupported, keyword)
	}
	c, err := draw(src)
	if err != nil {
		return "", err
	}
	return c.render(styles), nil
}

// diagramKeyword returns the first word of a diagram's declaration, such as
// "flowchart" in "flowchart LR", skipping the front matter, comments and
// blank lines that may come before it.
func diagramKeyword(src string) string {
	lines := strings.Split(src, "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "---"; i++ {
		}
		i++
	}
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		word, _, _ := strings.Cut(strings.Fields(line)[0], ";")
		return word
	}
	return ""
}
