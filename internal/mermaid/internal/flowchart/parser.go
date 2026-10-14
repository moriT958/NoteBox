package flowchart

import (
	"fmt"
	"notebox/internal/mermaid/internal/diagram"
	"notebox/internal/mermaid/internal/graph"
	"strings"
	"unicode"
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
	// open holds the subgraphs being read, the innermost last.
	open []*subgraph
	// groups maps the IDs of the subgraphs read so far to indexes into
	// g.Groups, and parents holds the group each group is in, or -1.
	groups  map[string]int
	parents []int
	// grouped reports for each node whether it is in a group.
	grouped []bool
}

// A subgraph being read, from its "subgraph" statement to its "end".
type subgraph struct {
	id, title string
	// nodes holds the nodes its statements name, and groups the groups that
	// ended in it.
	nodes, groups []int
}

// parseFlowchart parses the source of a flowchart, which starts with a
// "graph" or "flowchart" declaration. Errors tell the line the statement
// that can't be read is on.
func parseFlowchart(src string) (*graph.Graph, error) {
	stmts := splitStatements(src)
	if len(stmts) == 0 {
		return nil, fmt.Errorf("%w: no declaration", diagram.ErrSyntax)
	}
	p := &parser{g: &graph.Graph{}, index: map[string]int{}, groups: map[string]int{}}
	if err := p.parseDeclaration(stmts[0].text); err != nil {
		return nil, fmt.Errorf("line %d: %w", stmts[0].line, err)
	}
	for _, st := range stmts[1:] {
		if err := p.parse(st.text); err != nil {
			return nil, fmt.Errorf("line %d: %w", st.line, err)
		}
	}
	if len(p.open) > 0 {
		return nil, fmt.Errorf("%w: subgraph %q without an end", diagram.ErrSyntax, p.open[len(p.open)-1].title)
	}
	for id := range p.groups {
		if _, ok := p.index[id]; ok {
			return nil, fmt.Errorf("%w: link to subgraph %q", diagram.ErrUnsupported, id)
		}
	}
	return p.g, nil
}

// parse parses a statement after the declaration.
func (p *parser) parse(st string) error {
	fields := strings.Fields(st)
	switch keyword := fields[0]; {
	case keyword == "subgraph":
		p.open = append(p.open, parseSubgraph(strings.TrimSpace(strings.TrimPrefix(st, keyword))))
		return nil
	case keyword == "end" && len(fields) == 1:
		return p.endSubgraph()
	case ignoredStatements[keyword]:
		return nil
	}
	return p.parseStatement(st)
}

// parseSubgraph parses what follows "subgraph": nothing, an ID, an ID and
// a title in brackets, or a title. A title with spaces in it names no
// subgraph that other statements can refer to.
func parseSubgraph(rest string) *subgraph {
	if open := strings.IndexByte(rest, '['); open > 0 && strings.HasSuffix(rest, "]") {
		if id := strings.TrimSpace(rest[:open]); !strings.ContainsFunc(id, unicode.IsSpace) {
			return &subgraph{id: id, title: diagram.CleanLabel(rest[open+1 : len(rest)-1])}
		}
	}
	title := diagram.CleanLabel(rest)
	if strings.ContainsFunc(title, unicode.IsSpace) {
		return &subgraph{title: title}
	}
	return &subgraph{id: title, title: title}
}

// endSubgraph ends the innermost subgraph being read. As in Mermaid, it
// takes the nodes its statements name that aren't in a subgraph that ended
// before it, and a subgraph with an ID that ended before takes them
// instead. A subgraph without nodes in it is drawn as a node.
func (p *parser) endSubgraph() error {
	if len(p.open) == 0 {
		return fmt.Errorf("%w: end without a subgraph", diagram.ErrSyntax)
	}
	sg := p.open[len(p.open)-1]
	p.open = p.open[:len(p.open)-1]

	var nodes []int
	for _, n := range sg.nodes {
		if !p.grouped[n] {
			p.grouped[n] = true
			nodes = append(nodes, n)
		}
	}
	gi, ok := p.groups[sg.id]
	if !ok {
		if len(nodes) == 0 && len(sg.groups) == 0 {
			p.mention(p.addNode(graph.Node{Label: sg.title}))
			return nil
		}
		gi = len(p.g.Groups)
		p.g.Groups = append(p.g.Groups, graph.Group{Label: sg.title})
		p.parents = append(p.parents, -1)
		if sg.id != "" {
			p.groups[sg.id] = gi
		}
	}
	p.g.Groups[gi].Nodes = append(p.g.Groups[gi].Nodes, nodes...)
	for _, c := range sg.groups {
		if err := p.nest(c, gi); err != nil {
			return err
		}
	}
	if len(p.open) > 0 {
		outer := p.open[len(p.open)-1]
		outer.groups = append(outer.groups, gi)
	}
	return nil
}

// nest puts group c in group gi, unless c is in a group already.
func (p *parser) nest(c, gi int) error {
	if p.parents[c] >= 0 {
		return nil
	}
	for a := gi; a >= 0; a = p.parents[a] {
		if a == c {
			return fmt.Errorf("%w: subgraph %q in itself", diagram.ErrSyntax, p.g.Groups[c].Label)
		}
	}
	p.parents[c] = gi
	p.g.Groups[gi].Groups = append(p.g.Groups[gi].Groups, c)
	return nil
}

// mention adds node n to the innermost subgraph being read, if any.
func (p *parser) mention(n int) {
	if len(p.open) > 0 {
		sg := p.open[len(p.open)-1]
		sg.nodes = append(sg.nodes, n)
	}
}

func (p *parser) addNode(n graph.Node) int {
	p.g.Nodes = append(p.g.Nodes, n)
	p.grouped = append(p.grouped, false)
	return len(p.g.Nodes) - 1
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
		i = p.addNode(graph.Node{Label: label, Shape: sh})
		p.index[id] = i
	} else if hasLabel {
		p.g.Nodes[i].Label, p.g.Nodes[i].Shape = label, sh
	}
	p.mention(i)
	return i, nil
}
