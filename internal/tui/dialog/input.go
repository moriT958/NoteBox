package dialog

import (
	"image"
	"strings"

	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const NewNoteID = "new-note"

// Input asks for a single line of text.
type Input struct {
	com     *common.Common
	id      string
	title   string
	confirm string
	input   textinput.Model
	submit  func(value string) Action
	keys    struct{ Submit, Close key.Binding }
}

// NewNoteInput asks for the title of a new note.
func NewNoteInput(com *common.Common) *Input {
	return newInput(com, NewNoteID, "New Note", "Enter note name...", "Create",
		func(title string) Action { return ActionCreateNote{Title: title} })
}

func newInput(com *common.Common, id, title, placeholder, confirm string, submit func(string) Action) *Input {
	d := &Input{
		com:     com,
		id:      id,
		title:   title,
		confirm: confirm,
		input:   common.NewInput(placeholder, 100),
		submit:  submit,
	}
	d.input.Focus()
	d.keys.Submit = enterKey(strings.ToLower(confirm))
	d.keys.Close = CloseKey
	return d
}

func (d *Input) ID() string { return d.id }

func (d *Input) HandleMsg(msg tea.Msg) Action {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, d.keys.Close):
			return ActionClose{}
		case key.Matches(msg, d.keys.Submit):
			if strings.TrimSpace(d.input.Value()) == "" {
				return nil
			}
			return d.submit(d.input.Value())
		}
	}
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return cmdAction(cmd)
}

func (d *Input) Render(area image.Rectangle) (string, *tea.Cursor) {
	f := newFrame(d.com, area)
	d.input.SetWidth(max(1, f.innerWidth()-3))
	const inputLine = 2
	view := f.render(
		d.com.Styles.Dialog.Title.Render(d.title),
		"",
		d.input.View(),
		"",
		f.buttons(d.confirm, "Cancel"),
	)
	return view, f.cursor(common.InputCursorX(d.input), inputLine)
}

func (d *Input) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.Submit, d.keys.Close}
}

func (d *Input) FullHelp() [][]key.Binding {
	return [][]key.Binding{d.ShortHelp()}
}
