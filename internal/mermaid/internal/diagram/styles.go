package diagram

import "image/color"

// Styles are the colors of the parts of a diagram. With the zero value, a
// diagram is drawn without colors. A part without a color is drawn in the
// terminal's default color rather than in the color of the text around the
// diagram, so a colored diagram should give every part a color.
type Styles struct {
	// Border is the color of the outlines of nodes and groups.
	Border color.Color
	// Text is the color of the labels of nodes and groups.
	Text color.Color
	// Edge is the color of the lines of edges and their heads.
	Edge color.Color
	// EdgeLabel is the color of the labels of edges.
	EdgeLabel color.Color
}
