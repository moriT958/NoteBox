// Package keymap holds every key binding of the TUI and applies the user's
// overrides from the config.
package keymap

import (
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
)

// KeyMap holds every key binding, grouped by where it's active.
type KeyMap struct {
	Global struct {
		Quit, Help, Boxes key.Binding
	}
	List struct {
		Up, Down, New, Open, FocusPreview, Rename, Delete, Edit, Search key.Binding
	}
	Rename struct {
		Confirm, Cancel key.Binding
	}
	Preview struct {
		ScrollUp, ScrollDown, ScrollLeft, ScrollRight    key.Binding
		HalfPageUp, HalfPageDown, PageUp, PageDown       key.Binding
		Top, Bottom                                      key.Binding
		FocusList, Edit, Pin, CloseTab, NextTab, PrevTab key.Binding
	}
	// Dialog is shared by every dialog; a dialog may rename what Confirm
	// does in its help.
	Dialog struct {
		Up, Down, Confirm, Cancel key.Binding
	}
	Confirm struct {
		No key.Binding
	}
	Boxes struct {
		New, OpenFolder, Rename, Delete key.Binding
	}
	Form struct {
		Next, Prev key.Binding
	}
}

// action names a binding by its ID in the config.
type action struct {
	id string
	b  *key.Binding
}

// actions is the one list of every binding and its ID; applying overrides,
// finding conflicts, and the docs all go by it.
func (k *KeyMap) actions() []action {
	return []action{
		{"global.quit", &k.Global.Quit},
		{"global.help", &k.Global.Help},
		{"global.boxes", &k.Global.Boxes},

		{"list.up", &k.List.Up},
		{"list.down", &k.List.Down},
		{"list.new", &k.List.New},
		{"list.open", &k.List.Open},
		{"list.focus_preview", &k.List.FocusPreview},
		{"list.rename", &k.List.Rename},
		{"list.delete", &k.List.Delete},
		{"list.edit", &k.List.Edit},
		{"list.search", &k.List.Search},

		{"rename.confirm", &k.Rename.Confirm},
		{"rename.cancel", &k.Rename.Cancel},

		{"preview.scroll_up", &k.Preview.ScrollUp},
		{"preview.scroll_down", &k.Preview.ScrollDown},
		{"preview.scroll_left", &k.Preview.ScrollLeft},
		{"preview.scroll_right", &k.Preview.ScrollRight},
		{"preview.half_page_up", &k.Preview.HalfPageUp},
		{"preview.half_page_down", &k.Preview.HalfPageDown},
		{"preview.page_up", &k.Preview.PageUp},
		{"preview.page_down", &k.Preview.PageDown},
		{"preview.top", &k.Preview.Top},
		{"preview.bottom", &k.Preview.Bottom},
		{"preview.focus_list", &k.Preview.FocusList},
		{"preview.edit", &k.Preview.Edit},
		{"preview.pin", &k.Preview.Pin},
		{"preview.close_tab", &k.Preview.CloseTab},
		{"preview.next_tab", &k.Preview.NextTab},
		{"preview.prev_tab", &k.Preview.PrevTab},

		{"dialog.up", &k.Dialog.Up},
		{"dialog.down", &k.Dialog.Down},
		{"dialog.confirm", &k.Dialog.Confirm},
		{"dialog.cancel", &k.Dialog.Cancel},

		{"confirm.no", &k.Confirm.No},

		{"boxes.new", &k.Boxes.New},
		{"boxes.open_folder", &k.Boxes.OpenFolder},
		{"boxes.rename", &k.Boxes.Rename},
		{"boxes.delete", &k.Boxes.Delete},

		{"form.next", &k.Form.Next},
		{"form.prev", &k.Form.Prev},
	}
}

func Default() KeyMap {
	var k KeyMap

	k.Global.Quit = bind("quit", "ctrl+c")
	k.Global.Help = bind("more/less help", "?")
	k.Global.Boxes = bind("boxes", "ctrl+b")

	k.List.Up = bind("up", "up", "k", "ctrl+p")
	k.List.Down = bind("down", "down", "j", "ctrl+n")
	k.List.New = bind("new", "n")
	k.List.Open = bind("open tab", "enter")
	k.List.FocusPreview = bind("preview", "ctrl+l")
	k.List.Rename = bind("rename", "r")
	k.List.Delete = bind("delete", "d")
	k.List.Edit = bind("edit", "v")
	k.List.Search = bind("search", "/")

	k.Rename.Confirm = bind("confirm rename", "enter")
	k.Rename.Cancel = bind("cancel", "esc")

	// The preview's defaults follow less.
	k.Preview.ScrollUp = bind("up", "up", "k", "y", "ctrl+y", "ctrl+p")
	k.Preview.ScrollDown = bind("down", "down", "j", "e", "ctrl+e", "ctrl+n")
	k.Preview.ScrollLeft = bind("left", "left", "h")
	k.Preview.ScrollRight = bind("right", "right", "l")
	k.Preview.HalfPageUp = bind("½ page up", "u", "ctrl+u")
	k.Preview.HalfPageDown = bind("½ page down", "d", "ctrl+d")
	k.Preview.PageUp = bind("page up", "pgup", "b")
	k.Preview.PageDown = bind("page down", "pgdown", "space", "f", "ctrl+f")
	k.Preview.Top = bind("top", "g", "<", "home")
	k.Preview.Bottom = bind("bottom", "G", ">", "end")
	k.Preview.FocusList = bind("list", "ctrl+h")
	k.Preview.Edit = bind("edit", "v")
	k.Preview.Pin = bind("pin tab", "enter")
	k.Preview.CloseTab = bind("close tab", "w")
	k.Preview.NextTab = bind("next tab", "tab")
	k.Preview.PrevTab = bind("prev tab", "shift+tab")

	k.Dialog.Up = bind("up", "up", "ctrl+p")
	k.Dialog.Down = bind("down", "down", "ctrl+n")
	k.Dialog.Confirm = bind("confirm", "enter")
	k.Dialog.Cancel = bind("cancel", "esc")

	k.Confirm.No = bind("no", "esc", "q")

	k.Boxes.New = bind("new", "n")
	k.Boxes.OpenFolder = bind("open folder", "o")
	k.Boxes.Rename = bind("rename", "r")
	k.Boxes.Delete = bind("delete", "d")

	k.Form.Next = bind("next field", "tab", "down", "ctrl+n")
	k.Form.Prev = bind("prev field", "shift+tab", "up", "ctrl+p")

	return k
}

// Keys returns the keys of every action, by action ID.
func (k *KeyMap) Keys() map[string][]string {
	m := map[string][]string{}
	for _, a := range k.actions() {
		m[a.id] = slices.Clone(a.b.Keys())
	}
	return m
}

// Viewport returns the preview's scroll keys as a viewport key map.
func (k *KeyMap) Viewport() viewport.KeyMap {
	p := k.Preview
	return viewport.KeyMap{
		Up:           p.ScrollUp,
		Down:         p.ScrollDown,
		Left:         p.ScrollLeft,
		Right:        p.ScrollRight,
		HalfPageUp:   p.HalfPageUp,
		HalfPageDown: p.HalfPageDown,
		PageUp:       p.PageUp,
		PageDown:     p.PageDown,
	}
}

// WithDesc returns b described as desc in the help, for a shared binding
// that does something specific in one place.
func WithDesc(b key.Binding, desc string) key.Binding {
	b.SetHelp(b.Help().Key, desc)
	return b
}

func bind(desc string, keys ...string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(helpKey(keys), desc))
}

func setKeys(b *key.Binding, keys []string) {
	b.SetKeys(keys...)
	b.SetHelp(helpKey(keys), b.Help().Desc)
}
