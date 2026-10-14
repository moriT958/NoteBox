// Package flowchart reads Mermaid flowcharts, which are drawn as graphs.
package flowchart

import (
	"fmt"

	"notebox/internal/mermaid/internal/diagram"
	"notebox/internal/mermaid/internal/graph"
)

// maxEdges is Mermaid's own limit on the edges of a flowchart by default.
const maxEdges = 500

// Draw draws a flowchart from its source.
func Draw(src string) (*diagram.Canvas, error) {
	g, err := parseFlowchart(src)
	if err != nil {
		return nil, err
	}
	if len(g.Nodes) == 0 {
		return nil, fmt.Errorf("%w: no nodes", diagram.ErrSyntax)
	}
	c, err := graph.Draw(g)
	if err != nil {
		return nil, fmt.Errorf("failed to draw flowchart: %w", err)
	}
	return c, nil
}
