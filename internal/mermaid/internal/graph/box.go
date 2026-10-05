package graph

import (
	"notebox/internal/mermaid/internal/diagram"

	"github.com/charmbracelet/x/ansi"
)

// drawBox draws a box with its top left corner at (x, y), and lines of text
// centered in it.
func drawBox(c *diagram.Canvas, x, y, w, h int, sh Shape, lines []string) {
	right, bottom := x+w-1, y+h-1
	c.HLine(x, right, y, diagram.LineSolid, diagram.RoleBorder)
	c.HLine(x, right, bottom, diagram.LineSolid, diagram.RoleBorder)
	c.VLine(x, y, bottom, diagram.LineSolid, diagram.RoleBorder)
	c.VLine(right, y, bottom, diagram.LineSolid, diagram.RoleBorder)

	corners := [4]string{"┌", "┐", "└", "┘"}
	switch sh {
	case ShapeRound:
		corners = [4]string{"╭", "╮", "╰", "╯"}
	case ShapeDiamond:
		corners = [4]string{"╱", "╲", "╲", "╱"}
	}
	c.Set(x, y, corners[0], diagram.RoleBorder)
	c.Set(right, y, corners[1], diagram.RoleBorder)
	c.Set(x, bottom, corners[2], diagram.RoleBorder)
	c.Set(right, bottom, corners[3], diagram.RoleBorder)

	inner := w - 2
	for i, line := range lines {
		line = diagram.FitLabel(line, inner)
		c.Text(x+1+(inner-ansi.StringWidth(line))/2, y+1+i, line, diagram.RoleText)
	}
}
