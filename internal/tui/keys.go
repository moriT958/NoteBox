package tui

import (
	"notebox/internal/tui/keymap"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// helpKeys implements help.KeyMap from fixed binding lists.
type helpKeys struct {
	short []key.Binding
	full  [][]key.Binding
}

func (h helpKeys) ShortHelp() []key.Binding  { return h.short }
func (h helpKeys) FullHelp() [][]key.Binding { return h.full }

func listHelp(k *keymap.KeyMap) help.KeyMap {
	l, g := k.List, k.Global
	return helpKeys{
		short: []key.Binding{l.Up, l.Down, l.New, l.Open, l.Rename, l.Delete, l.Search, g.Boxes, g.Help, g.Quit},
		full: [][]key.Binding{
			{l.Up, l.Down, l.FocusPreview},
			{l.New, l.Open, l.Rename, l.Delete, l.Edit, l.Search},
			{g.Boxes, g.Help, g.Quit},
		},
	}
}

func renameHelp(k *keymap.KeyMap) help.KeyMap {
	b := []key.Binding{k.Rename.Confirm, k.Rename.Cancel}
	return helpKeys{short: b, full: [][]key.Binding{b}}
}

func previewHelp(k *keymap.KeyMap) help.KeyMap {
	p, g := k.Preview, k.Global
	return helpKeys{
		short: []key.Binding{p.ScrollUp, p.ScrollDown, p.HalfPageUp, p.HalfPageDown, p.FocusList, p.Edit, g.Help, g.Quit},
		full: [][]key.Binding{
			{p.ScrollUp, p.ScrollDown, p.ScrollLeft, p.ScrollRight, p.HalfPageUp, p.HalfPageDown},
			{p.Pin, p.CloseTab, p.NextTab, p.PrevTab},
			{p.FocusList, p.Edit},
			{g.Boxes, g.Help, g.Quit},
		},
	}
}
