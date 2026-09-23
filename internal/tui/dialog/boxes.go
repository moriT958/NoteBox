package dialog

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"notebox/internal/core/box"
	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const BoxesID = "boxes"

// BoxItem is a box as listed in the dialog.
type BoxItem struct {
	Box box.Box
	// Missing reports that the box's directory no longer exists.
	Missing bool
}

// Boxes lists the active boxes and lets the user switch to, create, rename
// or delete one. It opens empty and shows a loading state until SetBoxes.
type Boxes struct {
	com       *common.Common
	currentID string
	loading   bool
	items     []BoxItem
	cursor    int
	offset    int

	renaming bool
	input    textinput.Model

	keys struct {
		Select, Close, Up, Down, New, OpenFolder, Rename, Delete key.Binding
		ConfirmRename, CancelRename                              key.Binding
	}
}

func NewBoxes(com *common.Common, currentID string) *Boxes {
	d := &Boxes{
		com:       com,
		currentID: currentID,
		loading:   true,
		input:     common.NewInput("", 100),
	}
	d.keys.Select = enterKey("select")
	d.keys.Close = CloseKey
	d.keys.Up = upKey
	d.keys.Down = downKey
	d.keys.New = key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new"))
	d.keys.OpenFolder = key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open folder"))
	d.keys.Rename = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename"))
	d.keys.Delete = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete"))
	d.keys.ConfirmRename = enterKey("confirm rename")
	d.keys.CancelRename = CloseKey
	return d
}

func (d *Boxes) ID() string { return BoxesID }

// SetBoxes installs the loaded boxes and puts the cursor on selectID, or
// keeps it at the same index when selectID isn't listed.
func (d *Boxes) SetBoxes(items []BoxItem, selectID string) {
	d.loading = false
	d.items = items
	if i := slices.IndexFunc(items, func(it BoxItem) bool { return it.Box.ID == selectID }); i >= 0 {
		d.cursor = i
	}
	d.cursor = max(0, min(d.cursor, len(items)-1))
	d.offset = scroll(d.cursor, d.offset, listRows)
}

func (d *Boxes) selected() (box.Box, bool) {
	if d.loading || len(d.items) == 0 {
		return box.Box{}, false
	}
	return d.items[d.cursor].Box, true
}

func (d *Boxes) HandleMsg(msg tea.Msg) Action {
	if d.renaming {
		return d.handleRename(msg)
	}
	kmsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	switch {
	case key.Matches(kmsg, d.keys.Close):
		return ActionClose{}
	case key.Matches(kmsg, d.keys.Up):
		d.cursor = max(0, d.cursor-1)
		d.offset = scroll(d.cursor, d.offset, listRows)
	case key.Matches(kmsg, d.keys.Down):
		d.cursor = max(0, min(d.cursor+1, len(d.items)-1))
		d.offset = scroll(d.cursor, d.offset, listRows)
	case key.Matches(kmsg, d.keys.New):
		return ActionOpen{Dialog: NewBoxForm(d.com, BoxFormNew)}
	case key.Matches(kmsg, d.keys.OpenFolder):
		return ActionOpen{Dialog: NewBoxForm(d.com, BoxFormOpenFolder)}
	}

	b, ok := d.selected()
	if !ok {
		return nil
	}
	switch {
	case key.Matches(kmsg, d.keys.Select):
		return ActionSwitchBox{Box: b}
	case key.Matches(kmsg, d.keys.Rename):
		d.renaming = true
		d.input.SetValue(b.Title)
		d.input.CursorEnd()
		d.input.Focus()
	case key.Matches(kmsg, d.keys.Delete):
		if b.ID == d.currentID {
			return nil
		}
		message := fmt.Sprintf("Delete box '%s'?", b.Title)
		return ActionOpen{Dialog: NewConfirm(d.com, message, ActionDeleteBox{Box: b})}
	}
	return nil
}

func (d *Boxes) handleRename(msg tea.Msg) Action {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, d.keys.CancelRename):
			d.stopRename()
			return nil
		case key.Matches(msg, d.keys.ConfirmRename):
			d.stopRename()
			b, ok := d.selected()
			title := d.input.Value()
			if !ok || strings.TrimSpace(title) == "" || title == b.Title {
				return nil
			}
			return ActionRenameBox{Box: b, Title: title}
		}
	}
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return cmdAction(cmd)
}

func (d *Boxes) stopRename() {
	d.renaming = false
	d.input.Blur()
}

const renamePrefix = "  "

func (d *Boxes) Render(area image.Rectangle) (string, *tea.Cursor) {
	f := newFrame(d.com, area)
	w := f.innerWidth()
	d.input.SetWidth(max(1, w-len(renamePrefix)-3))
	sty := d.com.Styles.Dialog

	// The title and a blank line plus the frame take the rest.
	rows := max(1, min(listRows, area.Dy()-2-sty.Frame.GetVerticalFrameSize()))
	offset := scroll(d.cursor, d.offset, rows)

	var list strings.Builder
	switch {
	case d.loading:
		list.WriteString("  Loading...")
	case len(d.items) == 0:
		list.WriteString("  No boxes found")
	}
	for i := offset; i < min(offset+rows, len(d.items)); i++ {
		if i != offset {
			list.WriteString("\n")
		}
		if i == d.cursor && d.renaming {
			list.WriteString(renamePrefix)
			list.WriteString(d.input.View())
			continue
		}

		it := d.items[i]
		title := it.Box.Title
		if it.Missing {
			title = sty.Missing.Render(title)
		}
		if it.Box.ID == d.currentID {
			title += " *"
		}
		line := "  " + title
		if i == d.cursor {
			line = sty.Cursor.Render(line)
		}
		list.WriteString(ansi.Truncate(line, w, "..."))
	}

	const listLine = 2
	view := f.render(sty.Title.Render("Boxes"), "", list.String())
	if !d.renaming {
		return view, nil
	}
	x := len(renamePrefix) + common.InputCursorX(d.input)
	return view, f.cursor(x, listLine+d.cursor-offset)
}

func (d *Boxes) ShortHelp() []key.Binding {
	if d.renaming {
		return []key.Binding{d.keys.ConfirmRename, d.keys.CancelRename}
	}
	return []key.Binding{d.keys.Select, d.keys.New, d.keys.OpenFolder, d.keys.Rename, d.keys.Delete, d.keys.Close}
}

func (d *Boxes) FullHelp() [][]key.Binding {
	if d.renaming {
		return [][]key.Binding{d.ShortHelp()}
	}
	return [][]key.Binding{
		{d.keys.Up, d.keys.Down, d.keys.Select},
		{d.keys.New, d.keys.OpenFolder, d.keys.Rename, d.keys.Delete},
		{d.keys.Close},
	}
}
