package graph

import (
	"fmt"
	"notebox/internal/mermaid/diagram"
)

// Limits on the size of a layout, beyond which a graph isn't drawn.
const (
	// maxPoints limits the points edges pass through, which edges spanning
	// many ranks add many of.
	maxPoints = 5000
	// maxCells limits the size of a drawing, in cells.
	maxCells = 1 << 21
)

const (
	// labelWidth and labelLines bound the size of a node's wrapped label,
	// and labelPad is the space on each side of it in its box.
	labelWidth = 24
	labelLines = 4
	labelPad   = 1
	// gapY is the least depth of the gap between ranks, in rows of a
	// layout.
	gapY = 2
	// orderSweeps is how many times the ranks are reordered to reduce the
	// crossings of edges, and placeSweeps how many times the boxes are
	// moved toward the boxes they are joined to.
	orderSweeps = 8
	placeSweeps = 10
	// labelGap is the number of columns from the middle of a node to where
	// the label of the edges ending at it starts, leaving a space after the
	// head.
	labelGap = 2
	// loopHeight is the least height of a box with an edge to itself, which
	// runs out of its right side on one row and back in on the next.
	loopHeight = 4
	// loopLabelX is how many columns right of a box's left side the label
	// of its edges to itself starts, past the loop, when a graph flows
	// sideways and the loop is under the box.
	loopLabelX = 4
)

// A layout places the nodes of a graph on a grid, in ranks from the
// top down. An edge that spans more than one rank passes through a point,
// a vertex without a node, in each rank between its ends, so that every
// part of an edge joins vertices in adjacent ranks.
//
// A graph that flows sideways is laid out the same way, with the columns
// and the rows of its boxes swapped, and turned when it is drawn. Every
// position in a layout is as it would be in a graph that flows down.
type layout struct {
	orient orientation
	verts  []vertex
	// ranks holds the vertices of each rank from left to right.
	ranks [][]int
	// paths holds the vertices each edge passes through from the top down,
	// which is from its to end to its from end for an edge going back up.
	// Edges from a node to itself have no path.
	paths [][]int
	// back reports for each edge whether it goes back up.
	back []bool
	// parts holds the parts of the edges, each joining vertices in
	// adjacent ranks.
	parts []part
	// rankY and rankH are the top row and the height of each rank. gapTop
	// holds the top row of the gap below each rank but the last, where the
	// parts of edges run sideways in tracks, one row each.
	rankY, rankH, gapTop []int
	w, h                 int
}

type vertex struct {
	// node is the index of the vertex's node, or -1 for a point an edge
	// passes through, edge being the index of that edge.
	node, edge int
	rank       int
	// lines is the wrapped label of a node.
	lines []string
	// x, y, w and h are the box of a node, or the column and the rows a
	// point takes up. cx is the column at the middle of the box.
	x, y, w, h int
	cx         int
	// label labels the edges that end at a node, and backLabel the edges
	// going back up from it. They are written beside the heads above the
	// node, label to the right and backLabel to the left, or along the
	// edges when the graph flows sideways.
	label, backLabel string
	// loop reports whether a node has edges to itself, labelled loopLabel.
	loop      bool
	loopLabel string
}

func (v *vertex) isPoint() bool { return v.node < 0 }

// loopDX is how many columns right of its middle the edges from a node to
// itself turn, a column out from its right side.
func (v *vertex) loopDX() int { return v.w - v.w/2 + 1 }

// A part of an edge joins vertices in adjacent ranks. It leaves its from
// vertex at column fromX and reaches its to vertex at column toX, running
// sideways along a track in the gap between the ranks when they differ.
type part struct {
	edge, from, to int
	first, last    bool
	fromX, toX     int
	track          int
}

// layOut places a graph's nodes, unless the layout would be too large.
func layOut(g *Graph) (*layout, error) {
	back := backEdges(g)
	ranks := rankNodes(g, back)
	points := 0
	for _, lk := range links(g, back) {
		points += ranks[lk.to] - ranks[lk.from] - 1
	}
	if points > maxPoints {
		return nil, fmt.Errorf("%w: edges span %d ranks between their ends", diagram.ErrTooLarge, points)
	}

	l := &layout{back: back, orient: orientation{sideways: g.Dir == LeftRight || g.Dir == RightLeft}}
	l.addVertices(g, ranks)
	l.initOrder()
	l.reduceCrossings()
	l.placeX()
	l.addParts()
	l.placeY()
	if l.w*l.h > maxCells {
		return nil, fmt.Errorf("%w: %dx%d cells", diagram.ErrTooLarge, l.w, l.h)
	}
	return l, nil
}
