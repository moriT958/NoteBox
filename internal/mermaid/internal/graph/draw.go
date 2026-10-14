package graph

import (
	"notebox/internal/mermaid/internal/diagram"
)

// Draw lays out a graph and draws it, or reports diagram.ErrTooLarge for a
// graph too large to draw. Frames are drawn first, so that the edges that
// cross them join their outlines, and their titles last, over the edges.
func Draw(g *Graph) (*diagram.Canvas, error) {
	if err := check(g); err != nil {
		return nil, err
	}
	l, err := layOut(g)
	if err != nil {
		return nil, err
	}

	p := newPen(g.Dir, l)
	for _, f := range l.frames {
		p.frame(f)
	}
	for _, v := range l.verts {
		if !v.isPoint() {
			p.box(v, g.Nodes[v.node].Shape)
		}
	}
	drawEdges(p, g)
	drawLabels(p)
	drawLoops(p, g)
	for _, f := range l.frames {
		p.title(f)
	}
	return p.c, nil
}
