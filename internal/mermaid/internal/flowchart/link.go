package flowchart

import (
	"fmt"
	"notebox/internal/mermaid/internal/diagram"
	"notebox/internal/mermaid/internal/graph"
	"strings"
)

// parseLink parses an edge between two node groups, such as "-->",
// "-.->|label|", "== label ==>" or "~~~", leaving its ends to the caller.
func parseLink(s *scanner) (graph.Edge, error) {
	start := s.i
	if id := edgeID(s); id != "" {
		return graph.Edge{}, fmt.Errorf("%w: edge ID %q", diagram.ErrUnsupported, id)
	}

	var e graph.Edge
	if h, ok := linkHead(s.peek(0)); ok && s.peek(0) != '>' && strings.ContainsRune("-=.", s.peek(1)) {
		e.FromHead = h
		if s.peek(0) == '<' {
			e.FromHead = graph.HeadArrow
		}
		s.i++
	}

	switch s.peek(0) {
	case '~':
		n := countRun(s, '~')
		if n < 3 || e.FromHead != graph.HeadNone {
			return graph.Edge{}, fmt.Errorf("%w: bad link %q", diagram.ErrSyntax, string(s.rs[start:s.i]))
		}
		e.Invisible, e.Length = true, n-2
	case '-':
		if s.peek(1) == '.' {
			return parseDottedLink(s, e)
		}
		if err := parseSolidLink(s, &e, '-', diagram.LineSolid); err != nil {
			return graph.Edge{}, err
		}
	case '=':
		if err := parseSolidLink(s, &e, '=', diagram.LineThick); err != nil {
			return graph.Edge{}, err
		}
	default:
		return graph.Edge{}, fmt.Errorf("%w: expected a link at %q", diagram.ErrSyntax, s.rest())
	}
	err := parsePipeLabel(s, &e)
	return e, err
}

// edgeID returns the ID that names the link s is at, as "e1" in "e1@-->",
// or "" if the link has none.
func edgeID(s *scanner) string {
	j := s.i
	for j < len(s.rs) && s.rs[j] != '@' && isIDRune(s.rs[j], 0) {
		j++
	}
	if j == s.i || j+1 >= len(s.rs) || s.rs[j] != '@' || !strings.ContainsRune("-=~.<ox", s.rs[j+1]) {
		return ""
	}
	return string(s.rs[s.i:j])
}

func linkHead(r rune) (graph.Head, bool) {
	switch r {
	case '<', '>':
		return graph.HeadArrow, true
	case 'o':
		return graph.HeadCircle, true
	case 'x':
		return graph.HeadCross, true
	}
	return graph.HeadNone, false
}

// parseSolidLink parses a link of dashes or of equals signs: either a whole
// link such as "-->" or "===", or one with its label inside, such as
// "-- label -->".
func parseSolidLink(s *scanner, e *graph.Edge, c rune, style diagram.LineStyle) error {
	e.Style = style
	n := countRun(s, c)
	if n < 2 {
		return fmt.Errorf("%w: bad link at %q", diagram.ErrSyntax, s.rest())
	}
	if h, ok := linkHead(s.peek(0)); ok && s.peek(0) != '<' {
		s.i++
		e.ToHead, e.Length = h, n-1
		return nil
	}
	if n >= 3 {
		e.Length = n - 2
		return nil
	}

	// "--" or "==" opens a label that runs until the rest of the link,
	// which is at least two more dashes or equals signs and a head, or
	// three of them.
	text, end, err := findLinkEnd(s, func(rs []rune, i int) (int, bool) {
		j := i
		for j < len(rs) && rs[j] == c {
			j++
		}
		switch run := j - i; {
		case run >= 2 && j < len(rs) && strings.ContainsRune("xo>", rs[j]):
			return j + 1, true
		case run >= 3:
			return j, true
		}
		return 0, false
	})
	if err != nil {
		return err
	}
	e.Label = diagram.CleanLabel(text)
	e.Length = len(end) - 2
	if tail := end[len(end)-1]; tail != c {
		e.ToHead, _ = linkHead(tail)
	}
	return nil
}

// parseDottedLink parses a dotted link such as "-.->" or "-..-", or one with
// its label inside, such as "-. label .->".
func parseDottedLink(s *scanner, e graph.Edge) (graph.Edge, error) {
	e.Style = diagram.LineDotted
	s.i++ // the dash
	dots := countRun(s, '.')
	if s.peek(0) == '-' {
		s.i++
		e.Length = dots
		if h, ok := linkHead(s.peek(0)); ok && s.peek(0) != '<' {
			s.i++
			e.ToHead = h
		}
		err := parsePipeLabel(s, &e)
		return e, err
	}
	if dots != 1 {
		return graph.Edge{}, fmt.Errorf("%w: bad link at %q", diagram.ErrSyntax, s.rest())
	}

	// "-." opens a label that runs until the rest of the link, which is
	// dots, a dash and maybe a head, with a dash before them or not.
	text, end, err := findLinkEnd(s, func(rs []rune, i int) (int, bool) {
		j := i
		if j+1 < len(rs) && rs[j] == '-' && rs[j+1] == '.' {
			j++
		}
		dots := j
		for j < len(rs) && rs[j] == '.' {
			j++
		}
		if j == dots || j >= len(rs) || rs[j] != '-' {
			return 0, false
		}
		j++
		if j < len(rs) && strings.ContainsRune("xo>", rs[j]) {
			j++
		}
		return j, true
	})
	if err != nil {
		return graph.Edge{}, err
	}
	e.Label = diagram.CleanLabel(text)
	e.Length = strings.Count(string(end), ".")
	if h, ok := linkHead(end[len(end)-1]); ok {
		e.ToHead = h
	}
	return e, nil
}

// findLinkEnd finds where the end of a link with its label inside matches,
// returning the label before it and the end itself.
func findLinkEnd(s *scanner, match func(rs []rune, i int) (int, bool)) (string, []rune, error) {
	for i := s.i; i < len(s.rs); i++ {
		if j, ok := match(s.rs, i); ok {
			text, end := string(s.rs[s.i:i]), s.rs[i:j]
			s.i = j
			return text, end, nil
		}
	}
	return "", nil, fmt.Errorf("%w: unclosed link label at %q", diagram.ErrSyntax, s.rest())
}

// parsePipeLabel parses the label that may follow a link, as in "-->|label|".
func parsePipeLabel(s *scanner, e *graph.Edge) error {
	s.skipSpaces()
	if s.peek(0) != '|' {
		return nil
	}
	s.i++
	text, err := readUntil(s, []string{"|"})
	if err != nil {
		return err
	}
	e.Label = diagram.CleanLabel(text)
	return nil
}
