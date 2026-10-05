package flowchart

import (
	"fmt"
	"notebox/internal/mermaid/diagram"
	"strings"
	"unicode"
)

// A scanner reads a statement rune by rune.
type scanner struct {
	rs []rune
	i  int
}

func (s *scanner) peek(n int) rune {
	if s.i+n >= len(s.rs) {
		return 0
	}
	return s.rs[s.i+n]
}

func (s *scanner) skipSpaces() {
	for s.i < len(s.rs) && unicode.IsSpace(s.rs[s.i]) {
		s.i++
	}
}

func (s *scanner) done() bool { return s.i >= len(s.rs) }

func (s *scanner) hasPrefix(p string) bool {
	i := s.i
	for _, r := range p {
		if i >= len(s.rs) || s.rs[i] != r {
			return false
		}
		i++
	}
	return true
}

func (s *scanner) rest() string { return string(s.rs[s.i:]) }

// isIDRune reports whether r, followed by next, can be part of a node ID.
// A dash or an equals sign that starts an edge ends the ID before it.
func isIDRune(r, next rune) bool {
	switch {
	case r == '-':
		return next != '>' && next != '-' && next != '.'
	case r == '=':
		return next != '='
	case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r):
		return true
	}
	return strings.ContainsRune("!#$%'*+.?\\_/`", r)
}

// readUntil reads up to the first of closers and skips it. Text in double
// quotes may contain closers.
func readUntil(s *scanner, closers []string) (string, error) {
	start := s.i
	inQuotes := false
	for !s.done() {
		if s.peek(0) == '"' {
			inQuotes = !inQuotes
		} else if !inQuotes {
			for _, c := range closers {
				if s.hasPrefix(c) {
					text := string(s.rs[start:s.i])
					s.i += len([]rune(c))
					return text, nil
				}
			}
		}
		s.i++
	}
	return "", fmt.Errorf("%w: unclosed %q", diagram.ErrSyntax, closers[0])
}

func countRun(s *scanner, r rune) int {
	n := 0
	for s.peek(0) == r {
		s.i++
		n++
	}
	return n
}
