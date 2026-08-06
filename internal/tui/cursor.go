package tui

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

func centeredLinePos(frameWidth, frameHeight, numLines, lineIdx int, line string) (x, y int) {
	topPad := max(0, (frameHeight-numLines)/2)
	leftPad := max(0, (frameWidth-lipgloss.Width(line))/2)
	return leftPad, topPad + lineIdx
}

// borderSize returns how many columns/rows s's border occupies on the left
// and top, so callers don't hardcode the "+1" a single-line border adds.
func borderSize(s lipgloss.Style) (x, y int) {
	return s.GetBorderLeftSize(), s.GetBorderTopSize()
}

// paddingSize returns how many columns/rows s's padding occupies on the left
// and top.
func paddingSize(s lipgloss.Style) (x, y int) {
	return s.GetPaddingLeft(), s.GetPaddingTop()
}

// realCursorX recomputes the input's cursor column using the display width
// of the text before it, since textinput.Model.Cursor() derives X from the
// rune count instead, which undercounts full-width characters.
func realCursorX(in textinput.Model) int {
	value := []rune(in.Value())
	pos := min(in.Position(), len(value))
	return lipgloss.Width(in.Prompt) + lipgloss.Width(string(value[:pos]))
}
