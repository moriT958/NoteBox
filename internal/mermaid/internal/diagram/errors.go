package diagram

import "errors"

// The errors a diagram that isn't drawn is reported with, so that the
// caller can show its source instead.
var (
	// ErrUnsupported is returned for diagram types, and syntax within them,
	// that can't be drawn yet.
	ErrUnsupported = errors.New("mermaid: unsupported diagram")
	// ErrSyntax is returned for diagrams that can't be parsed.
	ErrSyntax = errors.New("mermaid: syntax error")
	// ErrTooLarge is returned for diagrams too large to draw.
	ErrTooLarge = errors.New("mermaid: diagram too large")
)
