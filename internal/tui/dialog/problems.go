package dialog

import (
	"fmt"
	"image"
	"strings"

	"notebox/internal/tui/common"
	"notebox/internal/tui/keymap"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const ProblemsID = "config-problems"

// maxProblemRows is the most lines of problems shown at once.
const maxProblemRows = 12

// Problems lists what was wrong with the config at startup.
type Problems struct {
	com      *common.Common
	problems []error
	// offset is the first visible line of the problems.
	offset int
	// rows is how many lines the last render showed, for scrolling.
	rows int
	keys struct{ Close, Cancel, Up, Down key.Binding }
}

func NewProblems(com *common.Common, problems []error) *Problems {
	d := &Problems{com: com, problems: problems, rows: maxProblemRows}
	d.keys.Close = keymap.WithDesc(com.Keys.Dialog.Confirm, "close")
	d.keys.Cancel = com.Keys.Dialog.Cancel
	d.keys.Up = com.Keys.Dialog.Up
	d.keys.Down = com.Keys.Dialog.Down
	return d
}

func (d *Problems) ID() string { return ProblemsID }

func (d *Problems) HandleMsg(msg tea.Msg) Action {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, d.keys.Close), key.Matches(msg, d.keys.Cancel):
			return ActionClose{}
		case key.Matches(msg, d.keys.Up):
			d.offset = max(0, d.offset-1)
		case key.Matches(msg, d.keys.Down):
			d.offset++
		}
	}
	return nil
}

func (d *Problems) Render(area image.Rectangle) (string, *tea.Cursor) {
	f := newFrame(d.com, area)
	sty := d.com.Styles.Dialog

	intro := ansi.Wrap(fmt.Sprintf("%d problem(s) found in the config. The settings concerned use their defaults.", len(d.problems)), f.innerWidth(), "")
	var lines []string
	for _, p := range d.problems {
		for i, line := range strings.Split(ansi.Wrap(p.Error(), max(1, f.innerWidth()-2), ""), "\n") {
			prefix := "  "
			if i == 0 {
				prefix = "- "
			}
			lines = append(lines, prefix+line)
		}
	}

	// The title, the intro, the button, and a blank line between each take
	// the rest, along with the frame.
	chrome := 1 + lipgloss.Height(intro) + 1 + 1 + 3 + sty.Frame.GetVerticalFrameSize()
	d.rows = max(1, min(maxProblemRows, area.Dy()-chrome))
	d.offset = min(d.offset, max(0, len(lines)-d.rows))
	end := min(len(lines), d.offset+d.rows)

	button := lipgloss.PlaceHorizontal(f.innerWidth(), lipgloss.Center, sty.Confirm.Render(" (enter) Close "))
	view := f.render(
		sty.Title.Render("Config problems"),
		"",
		intro,
		"",
		padRows(lines[d.offset:end], min(len(lines), d.rows)),
		"",
		button,
	)
	return view, nil
}

func (d *Problems) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.Up, d.keys.Down, d.keys.Close}
}

func (d *Problems) FullHelp() [][]key.Binding {
	return [][]key.Binding{d.ShortHelp()}
}
