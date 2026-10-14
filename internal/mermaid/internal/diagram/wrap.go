package diagram

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// labelBreakpoints are where words too long for a line may break, in
// addition to hyphens, so identifiers split between their segments.
const labelBreakpoints = "_./"

// WrapLabel breaks a cleaned label into lines no wider than width. A label
// that needs more than maxLines lines is cut short with an ellipsis.
func WrapLabel(label string, width, maxLines int) []string {
	var lines []string
	for para := range strings.SplitSeq(label, "\n") {
		for l := range strings.SplitSeq(ansi.Wrap(para, width, labelBreakpoints), "\n") {
			lines = append(lines, strings.TrimSpace(l))
		}
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := lines[maxLines-1]
		if ansi.StringWidth(last) >= width {
			last = ansi.Truncate(last, width-1, "")
		}
		lines[maxLines-1] = last + "…"
	}
	return lines
}

// FitLabel cuts a single-line label short with an ellipsis if it is wider
// than width.
func FitLabel(label string, width int) string {
	return ansi.Truncate(label, width, "…")
}
