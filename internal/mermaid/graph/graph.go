// Package graph lays out and draws diagrams of nodes joined by edges, such
// as flowcharts, from what they mean rather than how they are written.
package graph

import "notebox/internal/mermaid/diagram"

// A Graph is a diagram of nodes joined by edges.
type Graph struct {
	Dir   Direction
	Nodes []Node
	Edges []Edge
}

// A Direction is the way a graph flows, from the nodes edges leave to the
// nodes they reach.
type Direction uint8

const (
	TopDown Direction = iota
	BottomUp
	LeftRight
	RightLeft
)

// A Node is drawn as a box around its label.
type Node struct {
	Label string
	Shape Shape
}

// A Shape is the outline a node's box is drawn with.
type Shape uint8

const (
	ShapeRect Shape = iota
	ShapeRound
	ShapeDiamond
)

// An Edge joins the node From to the node To, which may be the same node.
type Edge struct {
	// From and To are indexes into the graph's Nodes.
	From, To int
	Label    string
	// FromHead and ToHead are the heads at the From and To ends.
	FromHead, ToHead Head
	Style            diagram.LineStyle
	// Invisible edges only take part in the layout.
	Invisible bool
	// Length is the number of ranks the edge spans at least, one or more.
	Length int
}

// A Head is what an edge ends with.
type Head uint8

const (
	HeadNone Head = iota
	HeadArrow
	HeadCircle
	HeadCross
)
