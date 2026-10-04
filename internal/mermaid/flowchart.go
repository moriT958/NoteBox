package mermaid

import (
	"fmt"
	"strings"
	"unicode"
)

// flowDirection is the direction a flowchart flows in.
type flowDirection uint8

const (
	flowTopDown flowDirection = iota
	flowBottomUp
	flowLeftRight
	flowRightLeft
)

// shape is the outline a node is drawn with. Mermaid's many node shapes
// are drawn as the closest of these.
type shape uint8

const (
	shapeRect shape = iota
	shapeRound
	shapeDiamond
)

// head is what an edge ends with.
type head uint8

const (
	headNone head = iota
	headArrow
	headCircle
	headCross
)

type node struct {
	id    string
	label string
	shape shape
}

type edge struct {
	from, to int // indexes into flowchart.nodes
	label    string
	// fromHead and toHead are the heads at the from and to ends.
	fromHead, toHead head
	style            lineStyle
	// invisible edges only take part in the layout.
	invisible bool
	// length is the number of ranks the edge spans at least, made longer
	// with extra dashes, dots or equals signs.
	length int
}

type flowchart struct {
	dir   flowDirection
	nodes []node
	edges []edge
	index map[string]int // node IDs to indexes into nodes
}

// drawFlowchart draws a flowchart from its source.
func drawFlowchart(src string) (*canvas, error) {
	fc, err := parseFlowchart(src)
	if err != nil {
		return nil, err
	}
	if len(fc.nodes) == 0 {
		return nil, fmt.Errorf("%w: no nodes", ErrSyntax)
	}
	l, err := layOut(fc)
	if err != nil {
		return nil, err
	}
	p := newPen(fc.dir, l)
	for _, v := range l.verts {
		if !v.isPoint() {
			p.box(v, fc.nodes[v.node].shape)
		}
	}
	drawEdges(p, fc)
	return p.c, nil
}

// ignoredStatements only style or link a flowchart, which is drawn
// without them.
var ignoredStatements = map[string]bool{
	"classDef": true, "class": true, "style": true, "linkStyle": true,
	"click": true, "direction": true,
}

// parseFlowchart parses the source of a flowchart, which starts with a
// "graph" or "flowchart" declaration.
func parseFlowchart(src string) (*flowchart, error) {
	stmts := splitStatements(src)
	if len(stmts) == 0 {
		return nil, fmt.Errorf("%w: no declaration", ErrSyntax)
	}
	fc := &flowchart{index: map[string]int{}}
	if err := fc.parseDeclaration(stmts[0]); err != nil {
		return nil, err
	}
	for _, st := range stmts[1:] {
		keyword := strings.Fields(st)[0]
		switch {
		case keyword == "subgraph" || keyword == "end":
			return nil, fmt.Errorf("%w: subgraph", ErrUnsupported)
		case ignoredStatements[keyword]:
			continue
		}
		if err := fc.parseStatement(st); err != nil {
			return nil, err
		}
	}
	return fc, nil
}

func (fc *flowchart) parseDeclaration(st string) error {
	fields := strings.Fields(st)
	if fields[0] != "graph" && fields[0] != "flowchart" {
		return fmt.Errorf("%w: not a flowchart: %q", ErrSyntax, fields[0])
	}
	if len(fields) == 1 {
		return nil
	}
	if len(fields) > 2 {
		return fmt.Errorf("%w: unexpected %q after the direction", ErrSyntax, fields[2])
	}
	switch fields[1] {
	case "TB", "TD", "v":
		fc.dir = flowTopDown
	case "BT", "^":
		fc.dir = flowBottomUp
	case "LR", ">":
		fc.dir = flowLeftRight
	case "RL", "<":
		fc.dir = flowRightLeft
	default:
		return fmt.Errorf("%w: unknown direction %q", ErrSyntax, fields[1])
	}
	return nil
}

// splitStatements splits a flowchart into statements, which end at line
// breaks and semicolons, dropping front matter, comments, accessibility
// titles and descriptions, and empty statements.
func splitStatements(src string) []string {
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

	var stmts []string
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if isAccessibility(line, "accTitle", ":") || isAccessibility(line, "accDescr", ":") {
			continue
		}
		if isAccessibility(line, "accDescr", "{") {
			for ; i < len(lines) && !strings.Contains(lines[i], "}"); i++ {
			}
			continue
		}
		var cur strings.Builder
		flush := func() {
			if s := strings.TrimSpace(cur.String()); s != "" {
				stmts = append(stmts, s)
			}
			cur.Reset()
		}
		inQuotes := false
	scan:
		for j := 0; j < len(line); j++ {
			c := line[j]
			switch {
			case c == '"':
				inQuotes = !inQuotes
			case inQuotes:
			case c == '%' && strings.HasPrefix(line[j:], "%%"):
				break scan
			case c == ';':
				flush()
				continue
			}
			cur.WriteByte(c)
		}
		flush()
	}
	return stmts
}

// isAccessibility reports whether line is an accessibility statement that
// starts with keyword and then sep, which runs to the end of the line, or,
// for a "{", to the closing "}".
func isAccessibility(line, keyword, sep string) bool {
	rest, ok := strings.CutPrefix(line, keyword)
	return ok && strings.HasPrefix(strings.TrimSpace(rest), sep)
}

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

// parseStatement parses a statement made of groups of nodes joined by
// edges, such as "A & B --> C[Label] -.-> D".
func (fc *flowchart) parseStatement(st string) error {
	s := &scanner{rs: []rune(st)}
	from, err := fc.parseNodeGroup(s)
	if err != nil {
		return err
	}
	for {
		s.skipSpaces()
		if s.done() {
			return nil
		}
		e, err := parseLink(s)
		if err != nil {
			return err
		}
		s.skipSpaces()
		to, err := fc.parseNodeGroup(s)
		if err != nil {
			return err
		}
		for _, f := range from {
			for _, t := range to {
				e.from, e.to = f, t
				if len(fc.edges) == maxEdges {
					return fmt.Errorf("%w: more than %d edges", ErrTooLarge, maxEdges)
				}
				fc.edges = append(fc.edges, e)
			}
		}
		from = to
	}
}

// parseNodeGroup parses nodes joined by "&".
func (fc *flowchart) parseNodeGroup(s *scanner) ([]int, error) {
	var group []int
	for {
		s.skipSpaces()
		n, err := fc.parseNode(s)
		if err != nil {
			return nil, err
		}
		group = append(group, n)
		s.skipSpaces()
		if s.peek(0) != '&' {
			return group, nil
		}
		s.i++
	}
}

// nodeShapes are the brackets that a node's label can be written in, with
// the longest openers first so that "((" is tried before "(".
var nodeShapes = []struct {
	open    string
	closers []string
	shape   shape
}{
	{"(((", []string{")))"}, shapeRound},
	{"((", []string{"))"}, shapeRound},
	{"([", []string{"])"}, shapeRound},
	{"(-", []string{"-)"}, shapeRound},
	{"[[", []string{"]]"}, shapeRect},
	{"[(", []string{")]"}, shapeRound},
	{"[/", []string{"/]", `\]`}, shapeRect},
	{`[\`, []string{`\]`, "/]"}, shapeRect},
	{"{{", []string{"}}"}, shapeDiamond},
	{"[", []string{"]"}, shapeRect},
	{"(", []string{")"}, shapeRound},
	{"{", []string{"}"}, shapeDiamond},
	{">", []string{"]"}, shapeRect},
}

// parseNode parses a node ID, the label and shape it may be given, and the
// class it may be given with ":::".
func (fc *flowchart) parseNode(s *scanner) (int, error) {
	start := s.i
	for !s.done() && isIDRune(s.peek(0), s.peek(1)) {
		s.i++
	}
	if s.i == start {
		if s.done() {
			return 0, fmt.Errorf("%w: missing node", ErrSyntax)
		}
		return 0, fmt.Errorf("%w: unexpected %q", ErrSyntax, s.rest())
	}
	id := string(s.rs[start:s.i])
	if s.peek(0) == '@' {
		return 0, fmt.Errorf("%w: shape data or edge ID on %q", ErrUnsupported, id)
	}

	label, sh, hasLabel := id, shapeRect, false
	for _, ns := range nodeShapes {
		if !s.hasPrefix(ns.open) {
			continue
		}
		s.i += len([]rune(ns.open))
		text, err := readUntil(s, ns.closers)
		if err != nil {
			return 0, fmt.Errorf("%w in node %q", err, id)
		}
		label, sh, hasLabel = cleanLabel(text), ns.shape, true
		break
	}

	if s.hasPrefix(":::") {
		s.i += 3
		for !s.done() && isIDRune(s.peek(0), s.peek(1)) {
			s.i++
		}
	}

	i, ok := fc.index[id]
	if !ok {
		i = len(fc.nodes)
		fc.index[id] = i
		fc.nodes = append(fc.nodes, node{id: id, label: label, shape: sh})
	} else if hasLabel {
		fc.nodes[i].label, fc.nodes[i].shape = label, sh
	}
	return i, nil
}

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
	return "", fmt.Errorf("%w: unclosed %q", ErrSyntax, closers[0])
}

// parseLink parses an edge between two node groups, such as "-->",
// "-.->|label|", "== label ==>" or "~~~", leaving its ends to the caller.
func parseLink(s *scanner) (edge, error) {
	start := s.i
	if id := edgeID(s); id != "" {
		return edge{}, fmt.Errorf("%w: edge ID %q", ErrUnsupported, id)
	}

	var e edge
	if h, ok := linkHead(s.peek(0)); ok && s.peek(0) != '>' && strings.ContainsRune("-=.", s.peek(1)) {
		e.fromHead = h
		if s.peek(0) == '<' {
			e.fromHead = headArrow
		}
		s.i++
	}

	switch s.peek(0) {
	case '~':
		n := countRun(s, '~')
		if n < 3 || e.fromHead != headNone {
			return edge{}, fmt.Errorf("%w: bad link %q", ErrSyntax, string(s.rs[start:s.i]))
		}
		e.invisible, e.length = true, n-2
	case '-':
		if s.peek(1) == '.' {
			return parseDottedLink(s, e)
		}
		if err := parseSolidLink(s, &e, '-', lineSolid); err != nil {
			return edge{}, err
		}
	case '=':
		if err := parseSolidLink(s, &e, '=', lineThick); err != nil {
			return edge{}, err
		}
	default:
		return edge{}, fmt.Errorf("%w: expected a link at %q", ErrSyntax, s.rest())
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

func linkHead(r rune) (head, bool) {
	switch r {
	case '<', '>':
		return headArrow, true
	case 'o':
		return headCircle, true
	case 'x':
		return headCross, true
	}
	return headNone, false
}

func countRun(s *scanner, r rune) int {
	n := 0
	for s.peek(0) == r {
		s.i++
		n++
	}
	return n
}

// parseSolidLink parses a link of dashes or of equals signs: either a whole
// link such as "-->" or "===", or one with its label inside, such as
// "-- label -->".
func parseSolidLink(s *scanner, e *edge, c rune, style lineStyle) error {
	e.style = style
	n := countRun(s, c)
	if n < 2 {
		return fmt.Errorf("%w: bad link at %q", ErrSyntax, s.rest())
	}
	if h, ok := linkHead(s.peek(0)); ok && s.peek(0) != '<' {
		s.i++
		e.toHead, e.length = h, n-1
		return nil
	}
	if n >= 3 {
		e.length = n - 2
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
	e.label = cleanLabel(text)
	e.length = len(end) - 2
	if tail := end[len(end)-1]; tail != c {
		e.toHead, _ = linkHead(tail)
	}
	return nil
}

// parseDottedLink parses a dotted link such as "-.->" or "-..-", or one with
// its label inside, such as "-. label .->".
func parseDottedLink(s *scanner, e edge) (edge, error) {
	e.style = lineDotted
	s.i++ // the dash
	dots := countRun(s, '.')
	if s.peek(0) == '-' {
		s.i++
		e.length = dots
		if h, ok := linkHead(s.peek(0)); ok && s.peek(0) != '<' {
			s.i++
			e.toHead = h
		}
		err := parsePipeLabel(s, &e)
		return e, err
	}
	if dots != 1 {
		return edge{}, fmt.Errorf("%w: bad link at %q", ErrSyntax, s.rest())
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
		return edge{}, err
	}
	e.label = cleanLabel(text)
	e.length = strings.Count(string(end), ".")
	if h, ok := linkHead(end[len(end)-1]); ok {
		e.toHead = h
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
	return "", nil, fmt.Errorf("%w: unclosed link label at %q", ErrSyntax, s.rest())
}

// parsePipeLabel parses the label that may follow a link, as in "-->|label|".
func parsePipeLabel(s *scanner, e *edge) error {
	s.skipSpaces()
	if s.peek(0) != '|' {
		return nil
	}
	s.i++
	text, err := readUntil(s, []string{"|"})
	if err != nil {
		return err
	}
	e.label = cleanLabel(text)
	return nil
}
