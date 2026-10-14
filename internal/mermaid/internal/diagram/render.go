package diagram

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Render returns the canvas as lines of text without trailing spaces, each
// part colored as styles says. Colors only set and reset the foreground, so
// that the attributes of the text around the diagram carry through it.
func (c *Canvas) Render(styles Styles) string {
	seqs := map[Role]string{}
	for r, col := range map[Role]color.Color{
		RoleBorder:    styles.Border,
		RoleText:      styles.Text,
		RoleEdge:      styles.Edge,
		RoleEdgeLabel: styles.EdgeLabel,
	} {
		if col != nil {
			seqs[r] = ansi.Style{}.ForegroundColor(col).String()
		}
	}
	reset := ansi.Style{}.ForegroundColor(nil).String()

	var b strings.Builder
	for y := range c.h {
		row := c.cells[y*c.w : (y+1)*c.w]
		end := len(row)
		for end > 0 && row[end-1].s == " " {
			end--
		}
		cur := ""
		for _, cl := range row[:end] {
			// Spaces show no color, so they don't need to switch to one.
			if cl.s != " " && cl.s != "" && seqs[cl.role] != cur {
				if cur != "" {
					b.WriteString(reset)
				}
				cur = seqs[cl.role]
				b.WriteString(cur)
			}
			b.WriteString(cl.s)
		}
		if cur != "" {
			b.WriteString(reset)
		}
		if y < c.h-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
