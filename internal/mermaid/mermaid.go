// Package mermaid draws Mermaid diagrams as text with box-drawing characters.
package mermaid

import (
	"errors"
	"fmt"
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

// A drawFunc draws a diagram of one type from its whole source.
type drawFunc func(src string) (string, error)

// diagrams maps the keyword that opens a diagram to the function drawing it.
var diagrams = map[string]drawFunc{}

// Render draws a Mermaid diagram. Diagrams that can't be drawn are reported
// with one of the errors above, so that the caller can show the source
// instead.
func Render(src string) (out string, err error) {
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
	return draw(src)
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
