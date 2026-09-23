package dialog

import (
	"image"
	"strings"

	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const BoxFormID = "box-form"

type BoxFormMode int

const (
	// BoxFormNew creates a new box directory.
	BoxFormNew BoxFormMode = iota
	// BoxFormOpenFolder registers an existing directory as a box.
	BoxFormOpenFolder
)

const (
	nameLabel = "Name:  "
	pathLabel = "Path:  "
)

// BoxForm asks for a box's name and path. It stays open until the root
// closes it on success, or reports a failure through SetError.
type BoxForm struct {
	com   *common.Common
	mode  BoxFormMode
	title textinput.Model
	path  textinput.Model
	// onPath reports that the path input has focus.
	onPath bool
	err    string
	keys   struct{ Submit, Close, Next, Prev key.Binding }
}

func NewBoxForm(com *common.Common, mode BoxFormMode) *BoxForm {
	d := &BoxForm{com: com, mode: mode}
	switch mode {
	case BoxFormNew:
		d.title = common.NewInput("Box name...", 100)
		d.path = common.NewInput("Base path (optional, e.g. ~/projects)...", 200)
		d.keys.Submit = enterKey("create")
	case BoxFormOpenFolder:
		d.title = common.NewInput("Display name...", 100)
		d.path = common.NewInput("Existing folder path (e.g. ~/projects/work)...", 200)
		d.keys.Submit = enterKey("open")
	}
	d.title.Focus()
	d.keys.Close = CloseKey
	d.keys.Next = key.NewBinding(key.WithKeys("tab", "down", "ctrl+n"), key.WithHelp("tab", "next field"))
	d.keys.Prev = key.NewBinding(key.WithKeys("shift+tab", "up", "ctrl+p"), key.WithHelp("shift+tab", "prev field"))
	return d
}

func (d *BoxForm) ID() string { return BoxFormID }

// SetError shows why creating the box failed.
func (d *BoxForm) SetError(err error) {
	d.err = err.Error()
}

func (d *BoxForm) HandleMsg(msg tea.Msg) Action {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, d.keys.Close):
			return ActionClose{}
		case key.Matches(msg, d.keys.Submit):
			return d.submit()
		case key.Matches(msg, d.keys.Next):
			d.focusPath(true)
			return nil
		case key.Matches(msg, d.keys.Prev):
			d.focusPath(false)
			return nil
		}
	}

	var cmd tea.Cmd
	if d.onPath {
		d.path, cmd = d.path.Update(msg)
	} else {
		d.title, cmd = d.title.Update(msg)
	}
	return cmdAction(cmd)
}

func (d *BoxForm) focusPath(onPath bool) {
	d.onPath = onPath
	if onPath {
		d.title.Blur()
		d.path.Focus()
	} else {
		d.path.Blur()
		d.title.Focus()
	}
}

func (d *BoxForm) submit() Action {
	title := strings.TrimSpace(d.title.Value())
	path := strings.TrimSpace(d.path.Value())
	if title == "" {
		d.err = "name is required"
		return nil
	}
	if d.mode == BoxFormOpenFolder {
		if path == "" {
			d.err = "path is required"
			return nil
		}
		return ActionOpenFolder{Title: title, Path: path}
	}
	return ActionCreateBox{Title: title, Path: path}
}

func (d *BoxForm) Render(area image.Rectangle) (string, *tea.Cursor) {
	f := newFrame(d.com, area)
	w := f.innerWidth()
	inputWidth := max(1, w-len(nameLabel)-3)
	d.title.SetWidth(inputWidth)
	d.path.SetWidth(inputWidth)
	sty := d.com.Styles.Dialog

	header, confirm := "New Box", "Create"
	if d.mode == BoxFormOpenFolder {
		header, confirm = "Open Folder as Box", "Open"
	}

	errLine := ""
	if d.err != "" {
		errLine = sty.Error.Render(ansi.Truncate(d.err, w, "..."))
	}

	const (
		titleLine = 2
		pathLine  = 4
	)
	view := f.render(
		sty.Title.Render(header),
		"",
		nameLabel+d.title.View(),
		"",
		pathLabel+d.path.View(),
		errLine,
		"",
		f.buttons(confirm, "Cancel"),
	)
	if d.onPath {
		return view, f.cursor(len(pathLabel)+common.InputCursorX(d.path), pathLine)
	}
	return view, f.cursor(len(nameLabel)+common.InputCursorX(d.title), titleLine)
}

func (d *BoxForm) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.Submit, d.keys.Next, d.keys.Close}
}

func (d *BoxForm) FullHelp() [][]key.Binding {
	return [][]key.Binding{{d.keys.Submit, d.keys.Next, d.keys.Prev, d.keys.Close}}
}
