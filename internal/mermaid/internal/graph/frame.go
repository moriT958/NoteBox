package graph

import (
	"fmt"
	"slices"
	"strings"

	"notebox/internal/mermaid/internal/diagram"

	"github.com/charmbracelet/x/ansi"
)

// titleWidth bounds the width of a group's label in the top of its frame.
const titleWidth = 40

// A frame is drawn around a group with nodes in it. Frames nest as their
// groups do, and a frame takes up the ranks from its top node down to its
// bottom node, as a rectangle that nodes outside it keep out of.
type frame struct {
	// parent is the index of the frame this one is nested in, or -1.
	parent   int
	children []int
	depth    int
	// title is the group's label fitted on one line.
	title string
	// top and bottom are the first and the last rank of the nodes in it.
	top, bottom int
	// key orders frames side by side in a rank, the same way in every rank
	// they share, so that they don't overlap.
	key float64
	// above reports whether edges into the frame's top have heads or
	// labels, which take the row above it.
	above bool
	// x, y, w and h are the frame's outline.
	x, y, w, h int
}

// check reports edges that don't join nodes or groups with nodes in them,
// and groups that don't nest: indexes out of range, nodes or groups in more
// than one group, and groups nested in themselves.
func check(g *Graph) error {
	nodeIn := make([]bool, len(g.Nodes))
	parent := make([]int, len(g.Groups))
	for i := range parent {
		parent[i] = -1
	}
	for i, gr := range g.Groups {
		for _, n := range gr.Nodes {
			if n < 0 || n >= len(g.Nodes) || nodeIn[n] {
				return fmt.Errorf("graph: group %d has node %d, which is out of range or in another group", i, n)
			}
			nodeIn[n] = true
		}
		for _, c := range gr.Groups {
			if c < 0 || c >= len(g.Groups) || parent[c] >= 0 {
				return fmt.Errorf("graph: group %d has group %d, which is out of range or in another group", i, c)
			}
			parent[c] = i
		}
	}
	for i := range g.Groups {
		// A group nested in itself is among its ancestors before the walk
		// up has taken a step for each group.
		for p, steps := parent[i], 0; p >= 0; p, steps = parent[p], steps+1 {
			if p == i || steps > len(g.Groups) {
				return fmt.Errorf("graph: group %d is nested in itself", i)
			}
		}
	}

	joins := func(i int, group bool) bool {
		if group {
			return i >= 0 && i < len(g.Groups) && hasNodes(g, i)
		}
		return i >= 0 && i < len(g.Nodes)
	}
	for i, e := range g.Edges {
		if !joins(e.From, e.FromGroup) || !joins(e.To, e.ToGroup) || e.Length < 1 {
			return fmt.Errorf("graph: edge %d from %d to %d of length %d doesn't join %d nodes or %d groups with nodes", i, e.From, e.To, e.Length, len(g.Nodes), len(g.Groups))
		}
	}
	return nil
}

// hasNodes reports whether a group has nodes in it or in the groups nested
// in it. The groups must nest.
func hasNodes(g *Graph, gi int) bool {
	if len(g.Groups[gi].Nodes) > 0 {
		return true
	}
	return slices.ContainsFunc(g.Groups[gi].Groups, func(c int) bool { return hasNodes(g, c) })
}

// addFrames adds a frame for each group with nodes in it, parents before
// the frames nested in them, and returns the frame each node and each group
// is in, or -1. The groups must nest.
func (l *layout) addFrames(g *Graph) (nodeFrame, groupFrame []int) {
	nodeFrame = make([]int, len(g.Nodes))
	for i := range nodeFrame {
		nodeFrame[i] = -1
	}
	groupFrame = make([]int, len(g.Groups))
	for i := range groupFrame {
		groupFrame[i] = -1
	}
	nested := make([]bool, len(g.Groups))
	for _, gr := range g.Groups {
		for _, c := range gr.Groups {
			nested[c] = true
		}
	}

	var add func(gi, parent int)
	add = func(gi, parent int) {
		if !hasNodes(g, gi) {
			return
		}
		f := len(l.frames)
		groupFrame[gi] = f
		fr := frame{parent: parent, title: groupTitle(g.Groups[gi].Label), top: -1, bottom: -1}
		if parent >= 0 {
			fr.depth = l.frames[parent].depth + 1
			l.frames[parent].children = append(l.frames[parent].children, f)
		}
		l.frames = append(l.frames, fr)
		for _, n := range g.Groups[gi].Nodes {
			nodeFrame[n] = f
		}
		for _, c := range g.Groups[gi].Groups {
			add(c, f)
		}
	}
	for gi := range g.Groups {
		if !nested[gi] {
			add(gi, -1)
		}
	}
	return nodeFrame, groupFrame
}

// spanFrames sets the ranks frames take up, from the ranks of the nodes in
// them.
func (l *layout) spanFrames(ranks, nodeFrame []int) {
	for n, f := range nodeFrame {
		if f >= 0 {
			l.frames[f].spanRank(ranks[n])
		}
	}
	// Nested frames widen the ranks of the frames around them.
	for _, fr := range slices.Backward(l.frames) {
		if p := fr.parent; p >= 0 {
			l.frames[p].spanRank(fr.top)
			l.frames[p].spanRank(fr.bottom)
		}
	}
}

func (f *frame) spanRank(r int) {
	if f.top < 0 || r < f.top {
		f.top = r
	}
	f.bottom = max(f.bottom, r)
}

// groupTitle fits the label of a group on one line.
func groupTitle(label string) string {
	return diagram.FitLabel(strings.ReplaceAll(label, "\n", " "), titleWidth)
}

// titleSpan is the least width of a frame with a title, which is written in
// its top as "┌─ title ─┐".
func (f *frame) titleSpan() int {
	if f.title == "" {
		return 0
	}
	return ansi.StringWidth(f.title) + 6
}

// lca returns the innermost frame that the frames a and b are both in, or
// -1 when they are only both in the whole graph.
func (l *layout) lca(a, b int) int {
	for a != b {
		if l.depth(a) >= l.depth(b) {
			a = l.frames[a].parent
		} else {
			b = l.frames[b].parent
		}
	}
	return a
}

func (l *layout) depth(f int) int {
	if f < 0 {
		return -1
	}
	return l.frames[f].depth
}

// within returns the frame nested directly in frame f that frame in is, or
// in is nested in. It returns -1 when in is f itself.
func (l *layout) within(f, in int) int {
	for in != f {
		p := l.frames[in].parent
		if p == f {
			return in
		}
		in = p
	}
	return -1
}

// pointFrame returns the frame that an edge from a node in frame top down to
// a node in frame bottom passes through at rank r. The edge stays in the
// frames around its top end down to their bottoms, and enters the frames
// around its bottom end at their tops, so that it crosses frames through
// their tops and bottoms rather than their sides where it can.
func (l *layout) pointFrame(top, bottom, r int) int {
	lca := l.lca(top, bottom)
	for f := top; f != lca; f = l.frames[f].parent {
		if l.frames[f].bottom >= r {
			return f
		}
	}
	for f := bottom; f != lca; f = l.frames[f].parent {
		if l.frames[f].top <= r {
			return f
		}
	}
	return lca
}

// frameSpans counts the ranks the frames take up, at least as many as the
// vertices that keep the frames' places in ranks without vertices in them.
func (l *layout) frameSpans() int {
	n := 0
	for _, f := range l.frames {
		n += f.bottom - f.top + 1
	}
	return n
}
