package preview

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// pos is a cell of the rendered note: a line and a column.
type pos struct {
	line, col int
}

func (a pos) before(b pos) bool {
	return a.line < b.line || a.line == b.line && a.col < b.col
}

// selection is the text selected with the mouse, from the cell where the
// button was pressed (anchor) to the cell under the pointer (head), both
// included. It's kept in cells of the rendered note rather than of the
// screen, so it stays on the same text while the preview scrolls.
type selection struct {
	anchor, head pos
	dragging     bool
}

func (s selection) empty() bool {
	return s.anchor == s.head
}

// bounds returns the first and the last selected cell.
func (s selection) bounds() (start, end pos) {
	if s.head.before(s.anchor) {
		return s.head, s.anchor
	}
	return s.anchor, s.head
}

// selectedText returns the plain text of lines from start to end, both
// included, as it's copied: trailing spaces are dropped, and so is the
// indentation the lines share, which is the margin glamour adds.
func selectedText(lines []string, start, end pos) string {
	var out []string
	for l := start.line; l <= end.line && l < len(lines); l++ {
		plain := ansi.Strip(lines[l])
		from, to := 0, ansi.StringWidth(plain)
		if l == start.line {
			from = start.col
		}
		if l == end.line {
			to = cellEnd(plain, end.col+1)
		}
		out = append(out, strings.TrimRight(ansi.Cut(plain, from, to), " "))
	}
	// A first line that starts mid-line has lost its indentation, so it
	// doesn't count toward the shared one.
	return strings.Join(dedent(out, start.col > 0), "\n")
}

// cellEnd returns col moved past a wide character that col would cut in
// half, so that a half-selected character is kept whole. ansi.Cut already
// keeps one cut by its left end.
func cellEnd(plain string, col int) int {
	if col < ansi.StringWidth(plain) && ansi.StringWidth(ansi.Truncate(plain, col, "")) < col {
		return col + 1
	}
	return col
}

// dedent removes the leading spaces that the non-blank lines share,
// leaving the first line out of the count if skipFirst is set.
func dedent(lines []string, skipFirst bool) []string {
	shared := -1
	for i, l := range lines {
		if l == "" || i == 0 && skipFirst {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " "))
		if shared < 0 || n < shared {
			shared = n
		}
	}
	if shared <= 0 {
		return lines
	}
	for i, l := range lines {
		n := min(shared, len(l)-len(strings.TrimLeft(l, " ")))
		lines[i] = l[n:]
	}
	return lines
}
