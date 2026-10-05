package graph

import (
	"errors"
	"testing"

	"notebox/internal/mermaid/internal/diagram"
)

func TestDrawInvalidEdges(t *testing.T) {
	for _, e := range []Edge{
		{From: 0, To: 2, Length: 1},
		{From: -1, To: 0, Length: 1},
		{From: 0, To: 1, Length: 0},
	} {
		g := &Graph{Nodes: []Node{{Label: "A"}, {Label: "B"}}, Edges: []Edge{e}}
		_, err := Draw(g)
		if err == nil || errors.Is(err, diagram.ErrSyntax) || errors.Is(err, diagram.ErrTooLarge) || errors.Is(err, diagram.ErrUnsupported) {
			t.Errorf("Draw() with edge %+v error = %v, want an error of its own", e, err)
		}
	}
}
