package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
)

// KeyMap holds every key outside dialogs, grouped by what has focus. Each
// dialog owns its own keys.
type KeyMap struct {
	List struct {
		Up, Down, New, Open, FocusPreview, Rename, Delete, Edit, Search key.Binding
	}
	Rename struct {
		Confirm, Cancel key.Binding
	}
	Preview struct {
		FocusList, Edit, Pin, CloseTab, NextTab, PrevTab key.Binding
	}

	Quit, Help, Boxes key.Binding
}

func DefaultKeyMap() KeyMap {
	var k KeyMap

	k.List.Up = key.NewBinding(key.WithKeys("up", "k", "ctrl+p"), key.WithHelp("↑/k", "up"))
	k.List.Down = key.NewBinding(key.WithKeys("down", "j", "ctrl+n"), key.WithHelp("↓/j", "down"))
	k.List.New = key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new"))
	k.List.Open = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open tab"))
	k.List.FocusPreview = key.NewBinding(key.WithKeys("right", "l", "ctrl+l"), key.WithHelp("→/l", "preview"))
	k.List.Rename = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename"))
	k.List.Delete = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete"))
	k.List.Edit = key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit"))
	k.List.Search = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search"))

	k.Rename.Confirm = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm rename"))
	k.Rename.Cancel = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

	k.Preview.FocusList = key.NewBinding(key.WithKeys("left", "h", "ctrl+h"), key.WithHelp("←/h", "list"))
	k.Preview.Edit = key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit"))
	k.Preview.Pin = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "pin tab"))
	k.Preview.CloseTab = key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "close tab"))
	k.Preview.NextTab = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab"))
	k.Preview.PrevTab = key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev tab"))

	k.Quit = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))
	k.Help = key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more/less help"))
	k.Boxes = key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "boxes"))

	return k
}

// helpKeys implements help.KeyMap from fixed binding lists.
type helpKeys struct {
	short []key.Binding
	full  [][]key.Binding
}

func (h helpKeys) ShortHelp() []key.Binding  { return h.short }
func (h helpKeys) FullHelp() [][]key.Binding { return h.full }

func (k KeyMap) listHelp() help.KeyMap {
	l := k.List
	return helpKeys{
		short: []key.Binding{l.Up, l.Down, l.New, l.Open, l.Rename, l.Delete, l.Search, k.Boxes, k.Help, k.Quit},
		full: [][]key.Binding{
			{l.Up, l.Down, l.FocusPreview},
			{l.New, l.Open, l.Rename, l.Delete, l.Edit, l.Search},
			{k.Boxes, k.Help, k.Quit},
		},
	}
}

func (k KeyMap) renameHelp() help.KeyMap {
	b := []key.Binding{k.Rename.Confirm, k.Rename.Cancel}
	return helpKeys{short: b, full: [][]key.Binding{b}}
}

func (k KeyMap) previewHelp(vp viewport.KeyMap) help.KeyMap {
	p := k.Preview
	return helpKeys{
		short: []key.Binding{vp.Up, vp.Down, vp.HalfPageUp, vp.HalfPageDown, p.FocusList, p.Edit, k.Help, k.Quit},
		full: [][]key.Binding{
			{vp.Up, vp.Down, vp.HalfPageUp, vp.HalfPageDown},
			{p.Pin, p.CloseTab, p.NextTab, p.PrevTab},
			{p.FocusList, p.Edit},
			{k.Boxes, k.Help, k.Quit},
		},
	}
}
