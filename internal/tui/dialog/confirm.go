package dialog

import (
	"image"

	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const ConfirmID = "confirm"

// Confirm asks a yes/no question and returns yes when confirmed.
type Confirm struct {
	com     *common.Common
	message string
	yes     Action
	keys    struct{ Yes, No key.Binding }
}

func NewConfirm(com *common.Common, message string, yes Action) *Confirm {
	d := &Confirm{com: com, message: message, yes: yes}
	d.keys.Yes = enterKey("yes")
	d.keys.No = key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc", "no"))
	return d
}

func (d *Confirm) ID() string { return ConfirmID }

func (d *Confirm) HandleMsg(msg tea.Msg) Action {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, d.keys.Yes):
			return d.yes
		case key.Matches(msg, d.keys.No):
			return ActionClose{}
		}
	}
	return nil
}

func (d *Confirm) Render(area image.Rectangle) (string, *tea.Cursor) {
	f := newFrame(d.com, area)
	message := lipgloss.NewStyle().Width(f.innerWidth()).Align(lipgloss.Center).Render(d.message)
	return f.render(message, "", f.buttons("Yes", "No")), nil
}

func (d *Confirm) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.Yes, d.keys.No}
}

func (d *Confirm) FullHelp() [][]key.Binding {
	return [][]key.Binding{d.ShortHelp()}
}
