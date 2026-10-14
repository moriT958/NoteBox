package diagram

// dirs is a set of directions in which a line leaves a cell.
type dirs uint8

const (
	dirUp dirs = 1 << iota
	dirDown
	dirLeft
	dirRight
)

var solidGlyphs = map[dirs]string{
	dirUp:                                "│",
	dirDown:                              "│",
	dirUp | dirDown:                      "│",
	dirLeft:                              "─",
	dirRight:                             "─",
	dirLeft | dirRight:                   "─",
	dirDown | dirRight:                   "┌",
	dirDown | dirLeft:                    "┐",
	dirUp | dirRight:                     "└",
	dirUp | dirLeft:                      "┘",
	dirUp | dirDown | dirRight:           "├",
	dirUp | dirDown | dirLeft:            "┤",
	dirDown | dirLeft | dirRight:         "┬",
	dirUp | dirLeft | dirRight:           "┴",
	dirUp | dirDown | dirLeft | dirRight: "┼",
}

var thickGlyphs = map[dirs]string{
	dirUp:                                "┃",
	dirDown:                              "┃",
	dirUp | dirDown:                      "┃",
	dirLeft:                              "━",
	dirRight:                             "━",
	dirLeft | dirRight:                   "━",
	dirDown | dirRight:                   "┏",
	dirDown | dirLeft:                    "┓",
	dirUp | dirRight:                     "┗",
	dirUp | dirLeft:                      "┛",
	dirUp | dirDown | dirRight:           "┣",
	dirUp | dirDown | dirLeft:            "┫",
	dirDown | dirLeft | dirRight:         "┳",
	dirUp | dirLeft | dirRight:           "┻",
	dirUp | dirDown | dirLeft | dirRight: "╋",
}

// lineGlyph returns the box-drawing character for a line leaving a cell in
// the directions d. Dotted lines only have straight characters, so their
// corners and junctions are solid.
func lineGlyph(d dirs, style LineStyle) string {
	switch style {
	case LineThick:
		return thickGlyphs[d]
	case LineDotted:
		switch d {
		case dirUp, dirDown, dirUp | dirDown:
			return "┆"
		case dirLeft, dirRight, dirLeft | dirRight:
			return "┄"
		}
	}
	return solidGlyphs[d]
}
