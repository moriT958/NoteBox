package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-runewidth"
)

const (
	ansiDim    = "\x1b[2m"
	ansiHeader = "\x1b[1;36m"
	ansiReset  = "\x1b[0m"
)

// printTable shows rows as an aligned table with a header when stdout is a
// terminal, and as plain tab-separated lines otherwise, so that pipes and
// scripts get something easy to split. Rows flagged in dimmed are shown dimmed
// on a terminal.
func printTable(header []string, rows [][]string, dimmed []bool) {
	writeTable(os.Stdout, isTerminal(os.Stdout), header, rows, dimmed)
}

func writeTable(w io.Writer, terminal bool, header []string, rows [][]string, dimmed []bool) {
	if !terminal {
		for _, row := range rows {
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
		return
	}

	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = runewidth.StringWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if l := runewidth.StringWidth(cell); l > widths[i] {
				widths[i] = l
			}
		}
	}

	line := func(cells []string) string {
		var b strings.Builder
		for i, cell := range cells {
			b.WriteString(cell)
			if i < len(cells)-1 {
				b.WriteString(strings.Repeat(" ", widths[i]-runewidth.StringWidth(cell)+2))
			}
		}
		return b.String()
	}

	fmt.Fprintln(w, ansiHeader+line(header)+ansiReset)
	for i, row := range rows {
		s := line(row)
		if i < len(dimmed) && dimmed[i] {
			s = ansiDim + s + ansiReset
		}
		fmt.Fprintln(w, s)
	}
}
