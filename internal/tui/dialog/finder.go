package dialog

import (
	"image"

	"notebox/internal/core/note"
	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"
)

const FinderID = "finder"

// listRows is how many results a list dialog shows at most.
const listRows = 10

// Finder fuzzy-searches notes by title.
type Finder struct {
	com      *common.Common
	input    textinput.Model
	notes    []note.Note
	filtered []note.Note
	cursor   int
	offset   int
	keys     struct{ Select, Close, Up, Down key.Binding }
}

func NewFinder(com *common.Common, notes []note.Note) *Finder {
	d := &Finder{
		com:      com,
		input:    common.NewInput("Search notes...", 50),
		notes:    notes,
		filtered: notes,
	}
	d.input.Focus()
	d.keys.Select = enterKey("select")
	d.keys.Close = CloseKey
	d.keys.Up = upKey
	d.keys.Down = downKey
	return d
}

func (d *Finder) ID() string { return FinderID }

func (d *Finder) HandleMsg(msg tea.Msg) Action {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, d.keys.Close):
			return ActionClose{}
		case key.Matches(msg, d.keys.Select):
			if len(d.filtered) == 0 {
				return nil
			}
			return ActionSelectNote{Note: d.filtered[d.cursor]}
		case key.Matches(msg, d.keys.Up):
			d.cursor = max(0, d.cursor-1)
			d.offset = scroll(d.cursor, d.offset, listRows)
			return nil
		case key.Matches(msg, d.keys.Down):
			d.cursor = max(0, min(d.cursor+1, len(d.filtered)-1))
			d.offset = scroll(d.cursor, d.offset, listRows)
			return nil
		}
	}

	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	d.filter()
	return cmdAction(cmd)
}

func (d *Finder) filter() {
	d.cursor, d.offset = 0, 0
	query := d.input.Value()
	if query == "" {
		d.filtered = d.notes
		return
	}
	titles := make([]string, len(d.notes))
	for i, n := range d.notes {
		titles[i] = n.Title()
	}
	matches := fuzzy.Find(query, titles)
	d.filtered = make([]note.Note, len(matches))
	for i, m := range matches {
		d.filtered[i] = d.notes[m.Index]
	}
}

func (d *Finder) Render(area image.Rectangle) (string, *tea.Cursor) {
	f := newFrame(d.com, area)
	w := f.innerWidth()
	d.input.SetWidth(max(1, w-3))

	// Title, input and two blank lines plus the frame take the rest.
	rows := max(1, min(listRows, area.Dy()-4-d.com.Styles.Dialog.Frame.GetVerticalFrameSize()))
	offset := scroll(d.cursor, d.offset, rows)

	lines := make([]string, 0, rows)
	if len(d.filtered) == 0 {
		lines = append(lines, "  No matches found")
	} else {
		for i := offset; i < min(offset+rows, len(d.filtered)); i++ {
			line := "  " + d.filtered[i].Title()
			if i == d.cursor {
				line = d.com.Styles.Dialog.Cursor.Render(line)
			}
			lines = append(lines, ansi.Truncate(line, w, "..."))
		}
	}

	const inputLine = 2
	view := f.render(
		d.com.Styles.Dialog.Title.Render("Find Note"),
		"",
		d.input.View(),
		"",
		padRows(lines, rows),
	)
	return view, f.cursor(common.InputCursorX(d.input), inputLine)
}

func (d *Finder) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.Up, d.keys.Down, d.keys.Select, d.keys.Close}
}

func (d *Finder) FullHelp() [][]key.Binding {
	return [][]key.Binding{d.ShortHelp()}
}
