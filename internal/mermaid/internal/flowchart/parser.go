package flowchart

import (
	"fmt"
	"notebox/internal/mermaid/internal/diagram"
	"notebox/internal/mermaid/internal/graph"
	"strings"
)

// ignoredStatements only style or link a flowchart, which is drawn
// without them.
var ignoredStatements = map[string]bool{
	"classDef": true, "class": true, "style": true, "linkStyle": true,
	"click": true, "direction": true,
}

// A parser reads the source of a flowchart into the graph it draws.
type parser struct {
	g *graph.Graph
	// index maps the IDs of the nodes read so far to indexes into g.Nodes.
	index map[string]int
}

// parseFlowchart parses the source of a flowchart, which starts with a
// "graph" or "flowchart" declaration. Errors tell the line the statement
// that can't be read is on.
func parseFlowchart(src string) (*graph.Graph, error) {
	stmts := splitStatements(src)
	if len(stmts) == 0 {
		return nil, fmt.Errorf("%w: no declaration", diagram.ErrSyntax)
	}
	p := &parser{g: &graph.Graph{}, index: map[string]int{}}
	if err := p.parseDeclaration(stmts[0].text); err != nil {
		return nil, fmt.Errorf("line %d: %w", stmts[0].line, err)
	}
	for _, st := range stmts[1:] {
		if err := p.parse(st.text); err != nil {
			return nil, fmt.Errorf("line %d: %w", st.line, err)
		}
	}
	return p.g, nil
}

// parse parses a statement after the declaration.
func (p *parser) parse(st string) error {
	keyword := strings.Fields(st)[0]
	switch {
	case keyword == "subgraph" || keyword == "end":
		return fmt.Errorf("%w: subgraph", diagram.ErrUnsupported)
	case ignoredStatements[keyword]:
		return nil
	}
	return p.parseStatement(st)
}

func (p *parser) parseDeclaration(st string) error {
	fields := strings.Fields(st)
	if fields[0] != "graph" && fields[0] != "flowchart" {
		return fmt.Errorf("%w: not a flowchart: %q", diagram.ErrSyntax, fields[0])
	}
	if len(fields) == 1 {
		return nil
	}
	if len(fields) > 2 {
		return fmt.Errorf("%w: unexpected %q after the direction", diagram.ErrSyntax, fields[2])
	}
	switch fields[1] {
	case "TB", "TD", "v":
		p.g.Dir = graph.TopDown
	case "BT", "^":
		p.g.Dir = graph.BottomUp
	case "LR", ">":
		p.g.Dir = graph.LeftRight
	case "RL", "<":
		p.g.Dir = graph.RightLeft
	default:
		return fmt.Errorf("%w: unknown direction %q", diagram.ErrSyntax, fields[1])
	}
	return nil
}

// parseStatement parses a statement made of groups of nodes joined by
// edges, such as "A & B --> C[Label] -.-> D".
func (p *parser) parseStatement(st string) error {
	s := &scanner{rs: []rune(st)}
	from, err := p.parseNodeGroup(s)
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
		to, err := p.parseNodeGroup(s)
		if err != nil {
			return err
		}
		for _, f := range from {
			for _, t := range to {
				e.From, e.To = f, t
				if len(p.g.Edges) == maxEdges {
					return fmt.Errorf("%w: more than %d edges", diagram.ErrTooLarge, maxEdges)
				}
				p.g.Edges = append(p.g.Edges, e)
			}
		}
		from = to
	}
}

// parseNodeGroup parses nodes joined by "&".
func (p *parser) parseNodeGroup(s *scanner) ([]int, error) {
	var group []int
	for {
		s.skipSpaces()
		n, err := p.parseNode(s)
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
	shape   graph.Shape
}{
	{"(((", []string{")))"}, graph.ShapeRound},
	{"((", []string{"))"}, graph.ShapeRound},
	{"([", []string{"])"}, graph.ShapeRound},
	{"(-", []string{"-)"}, graph.ShapeRound},
	{"[[", []string{"]]"}, graph.ShapeRect},
	{"[(", []string{")]"}, graph.ShapeRound},
	{"[/", []string{"/]", `\]`}, graph.ShapeRect},
	{`[\`, []string{`\]`, "/]"}, graph.ShapeRect},
	{"{{", []string{"}}"}, graph.ShapeDiamond},
	{"[", []string{"]"}, graph.ShapeRect},
	{"(", []string{")"}, graph.ShapeRound},
	{"{", []string{"}"}, graph.ShapeDiamond},
	{">", []string{"]"}, graph.ShapeRect},
}

// parseNode parses a node ID, the label and shape it may be given, and the
// class it may be given with ":::".
func (p *parser) parseNode(s *scanner) (int, error) {
	start := s.i
	for !s.done() && isIDRune(s.peek(0), s.peek(1)) {
		s.i++
	}
	if s.i == start {
		if s.done() {
			return 0, fmt.Errorf("%w: missing node", diagram.ErrSyntax)
		}
		return 0, fmt.Errorf("%w: unexpected %q", diagram.ErrSyntax, s.rest())
	}
	id := string(s.rs[start:s.i])
	if s.peek(0) == '@' {
		return 0, fmt.Errorf("%w: shape data or edge ID on %q", diagram.ErrUnsupported, id)
	}

	label, sh, hasLabel := id, graph.ShapeRect, false
	for _, ns := range nodeShapes {
		if !s.hasPrefix(ns.open) {
			continue
		}
		s.i += len([]rune(ns.open))
		text, err := readUntil(s, ns.closers)
		if err != nil {
			return 0, fmt.Errorf("%w in node %q", err, id)
		}
		label, sh, hasLabel = diagram.CleanLabel(text), ns.shape, true
		break
	}

	if s.hasPrefix(":::") {
		s.i += 3
		for !s.done() && isIDRune(s.peek(0), s.peek(1)) {
			s.i++
		}
	}

	i, ok := p.index[id]
	if !ok {
		i = len(p.g.Nodes)
		p.index[id] = i
		p.g.Nodes = append(p.g.Nodes, graph.Node{Label: label, Shape: sh})
	} else if hasLabel {
		p.g.Nodes[i].Label, p.g.Nodes[i].Shape = label, sh
	}
	return i, nil
}
