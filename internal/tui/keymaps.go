package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
)

type keyMap struct {
	// Global Key Maps
	quit       key.Binding
	toggleHelp key.Binding

	listPanel   listPanelKeyMap
	previewer   previewerKeyMap
	typingModal modalKeyMap
	warnModal   modalKeyMap
	fuzzyModal  fuzzyModalKeyMap
}

type listPanelKeyMap struct {
	up           key.Binding
	down         key.Binding
	newNote      key.Binding
	focusPreview key.Binding
	deleteNote   key.Binding
	editNote     key.Binding
	search       key.Binding
}

type previewerKeyMap struct {
	focusList    key.Binding
	editNote     key.Binding
	up           key.Binding
	down         key.Binding
	halfPageUp   key.Binding
	halfPageDown key.Binding
}

type modalKeyMap struct {
	confirm key.Binding
	cancel  key.Binding
}

type fuzzyModalKeyMap struct {
	confirm key.Binding
	cancel  key.Binding
	up      key.Binding
	down    key.Binding
}

func defaultKeyMap() keyMap {
	vpKeys := viewport.DefaultKeyMap()

	return keyMap{
		quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		toggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		listPanel: listPanelKeyMap{
			up: key.NewBinding(
				key.WithKeys("k"),
				key.WithHelp("k", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("j"),
				key.WithHelp("j", "down"),
			),
			newNote: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new"),
			),
			focusPreview: key.NewBinding(
				key.WithKeys("ctrl+l"),
				key.WithHelp("ctrl+l", "preview"),
			),
			deleteNote: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
			editNote: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			search: key.NewBinding(
				key.WithKeys("/"),
				key.WithHelp("/", "search"),
			),
		},
		previewer: previewerKeyMap{
			focusList: key.NewBinding(
				key.WithKeys("ctrl+h"),
				key.WithHelp("ctrl+h", "list"),
			),
			editNote: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			up:           vpKeys.Up,
			down:         vpKeys.Down,
			halfPageUp:   vpKeys.HalfPageUp,
			halfPageDown: vpKeys.HalfPageDown,
		},
		typingModal: modalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "create"),
			),
			cancel: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "cancel"),
			),
		},
		warnModal: modalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "yes"),
			),
			cancel: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "no"),
			),
		},
		fuzzyModal: fuzzyModalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "select"),
			),
			cancel: key.NewBinding(
				key.WithKeys("ctrl+c", "esc"),
				key.WithHelp("ctrl+c/esc", "cancel"),
			),
			up: key.NewBinding(
				key.WithKeys("ctrl+p", "up"),
				key.WithHelp("ctrl+p/up", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("ctrl+n", "down"),
				key.WithHelp("ctrl+n/down", "down"),
			),
		},
	}
}

type focusedKeyMap struct {
	keys  keyMap
	focus focus
}

var _ help.KeyMap = focusedKeyMap{}

func (k keyMap) forFocus(f focus) focusedKeyMap {
	return focusedKeyMap{keys: k, focus: f}
}

func (k focusedKeyMap) ShortHelp() []key.Binding {
	switch k.focus {
	case onListPanel:
		return []key.Binding{
			k.keys.listPanel.up,
			k.keys.listPanel.down,
			k.keys.listPanel.newNote,
			k.keys.listPanel.deleteNote,
			k.keys.listPanel.search,
			k.keys.toggleHelp,
			k.keys.quit,
		}
	case onPreviewer:
		return []key.Binding{
			k.keys.previewer.up,
			k.keys.previewer.down,
			k.keys.previewer.halfPageUp,
			k.keys.previewer.halfPageDown,
			k.keys.previewer.focusList,
			k.keys.previewer.editNote,
			k.keys.toggleHelp,
			k.keys.quit,
		}
	default:
		return []key.Binding{k.keys.toggleHelp, k.keys.quit}
	}
}

func (k focusedKeyMap) FullHelp() [][]key.Binding {
	switch k.focus {
	case onListPanel:
		return [][]key.Binding{
			{k.keys.listPanel.up, k.keys.listPanel.down, k.keys.listPanel.focusPreview},
			{k.keys.listPanel.newNote, k.keys.listPanel.deleteNote, k.keys.listPanel.editNote, k.keys.listPanel.search},
			{k.keys.toggleHelp, k.keys.quit},
		}
	case onPreviewer:
		return [][]key.Binding{
			{k.keys.previewer.up, k.keys.previewer.down},
			{k.keys.previewer.halfPageUp, k.keys.previewer.halfPageDown},
			{k.keys.previewer.focusList, k.keys.previewer.editNote},
			{k.keys.toggleHelp, k.keys.quit},
		}
	default:
		return [][]key.Binding{{k.keys.toggleHelp, k.keys.quit}}
	}
}
