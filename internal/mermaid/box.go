package mermaid

// drawBox draws a box with its top left corner at (x, y), and lines of text
// centered in it.
func drawBox(c *canvas, x, y, w, h int, sh shape, lines []string) {
	right, bottom := x+w-1, y+h-1
	c.hline(x, right, y, lineSolid, roleBorder)
	c.hline(x, right, bottom, lineSolid, roleBorder)
	c.vline(x, y, bottom, lineSolid, roleBorder)
	c.vline(right, y, bottom, lineSolid, roleBorder)

	corners := [4]string{"┌", "┐", "└", "┘"}
	switch sh {
	case shapeRound:
		corners = [4]string{"╭", "╮", "╰", "╯"}
	case shapeDiamond:
		corners = [4]string{"╱", "╲", "╲", "╱"}
	}
	c.set(x, y, corners[0], roleBorder)
	c.set(right, y, corners[1], roleBorder)
	c.set(x, bottom, corners[2], roleBorder)
	c.set(right, bottom, corners[3], roleBorder)

	inner := w - 2
	for i, line := range lines {
		line = fitLabel(line, inner)
		c.text(x+1+(inner-textWidth(line))/2, y+1+i, line, roleText)
	}
}
