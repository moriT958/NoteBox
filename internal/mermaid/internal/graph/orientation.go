package graph

// An orientation holds the dimensions of a layout that depend on the way its
// graph flows. Layouts are always laid out as if they flowed down; a graph
// that flows sideways has the columns and rows of its boxes swapped, and
// needs other gaps between them.
type orientation struct {
	// sideways reports whether the graph flows left or right.
	sideways bool
}

// labelsAlong reports whether edge labels run along the edges, in the gaps
// between ranks, rather than beside the heads above the nodes. Labels can't
// be written down a column, so they do when the graph flows sideways.
func (o orientation) labelsAlong() bool { return o.sideways }

// gapX is the number of columns between vertices side by side in a rank: a
// few columns, or a row when a graph flows sideways.
func (o orientation) gapX() int {
	if o.sideways {
		return 1
	}
	return 3
}

// backX is the column, left of the middle of a node, that edges going back
// up leave it or reach it at, so that they keep apart from the edges going
// down: two columns, or a row when a graph flows sideways. backWidth is
// the least width of a box with them, keeping the column off its corners.
func (o orientation) backX() int {
	if o.sideways {
		return 1
	}
	return 2
}

func (o orientation) backWidth() int { return 2*o.backX() + 3 }
