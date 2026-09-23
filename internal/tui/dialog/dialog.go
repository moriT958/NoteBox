package dialog

import (
	"image"
	"slices"
	"strings"

	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// maxWidth is the widest a dialog gets, frame included.
const maxWidth = 60

type Dialog interface {
	ID() string
	HandleMsg(msg tea.Msg) Action
	// Render returns the dialog sized to fit area, and the cursor position
	// relative to the dialog's top-left corner (nil for no cursor).
	Render(area image.Rectangle) (string, *tea.Cursor)
	help.KeyMap
}

// Overlay is the stack of open dialogs; only the front one gets input.
type Overlay struct {
	dialogs []Dialog
}

func NewOverlay() *Overlay {
	return &Overlay{}
}

func (o *Overlay) HasDialogs() bool {
	return len(o.dialogs) > 0
}

func (o *Overlay) Open(d Dialog) {
	o.dialogs = append(o.dialogs, d)
}

func (o *Overlay) CloseFront() {
	if len(o.dialogs) > 0 {
		o.dialogs = o.dialogs[:len(o.dialogs)-1]
	}
}

func (o *Overlay) Close(id string) {
	o.dialogs = slices.DeleteFunc(o.dialogs, func(d Dialog) bool { return d.ID() == id })
}

// Front returns the dialog receiving input, or nil.
func (o *Overlay) Front() Dialog {
	if len(o.dialogs) == 0 {
		return nil
	}
	return o.dialogs[len(o.dialogs)-1]
}

// Find returns the open dialog with the given ID, or nil.
func (o *Overlay) Find(id string) Dialog {
	for _, d := range o.dialogs {
		if d.ID() == id {
			return d
		}
	}
	return nil
}

// Dialogs returns the open dialogs from back to front.
func (o *Overlay) Dialogs() []Dialog {
	return o.dialogs
}

func (o *Overlay) HandleMsg(msg tea.Msg) Action {
	if d := o.Front(); d != nil {
		return d.HandleMsg(msg)
	}
	return nil
}

// Key bindings shared by the dialogs.
var (
	CloseKey = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
	upKey    = key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("↑", "up"))
	downKey  = key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("↓", "down"))
)

func enterKey(desc string) key.Binding {
	return key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", desc))
}

// frame lays out a dialog: its size, and the frame drawn around its body.
type frame struct {
	com   *common.Common
	width int
}

func newFrame(com *common.Common, area image.Rectangle) frame {
	return frame{com: com, width: max(0, min(maxWidth, area.Dx()))}
}

// innerWidth is the width available to the body.
func (f frame) innerWidth() int {
	return max(0, f.width-f.com.Styles.Dialog.Frame.GetHorizontalFrameSize())
}

func (f frame) render(lines ...string) string {
	return f.com.Styles.Dialog.Frame.Width(f.width).Render(strings.Join(lines, "\n"))
}

// cursor converts a position within the body into one relative to the
// dialog's top-left corner.
func (f frame) cursor(x, bodyLine int) *tea.Cursor {
	s := f.com.Styles.Dialog.Frame
	return tea.NewCursor(
		s.GetBorderLeftSize()+s.GetPaddingLeft()+x,
		s.GetBorderTopSize()+s.GetPaddingTop()+bodyLine,
	)
}

// buttons renders the confirm/cancel buttons centered in the body.
func (f frame) buttons(confirm, cancel string) string {
	s := f.com.Styles.Dialog
	row := s.Confirm.Render(" (enter) "+confirm+" ") + "     " + s.Cancel.Render(" (esc) "+cancel+" ")
	return lipgloss.PlaceHorizontal(f.innerWidth(), lipgloss.Center, row)
}

// padRows joins lines into a block of exactly rows lines, padding with
// blank lines so a list dialog's rendered height stays constant regardless
// of how many lines currently have content.
func padRows(lines []string, rows int) string {
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// scroll keeps a list's cursor inside the rows visible from offset and
// returns the adjusted offset.
func scroll(cursor, offset, rows int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+rows {
		return cursor - rows + 1
	}
	return offset
}
