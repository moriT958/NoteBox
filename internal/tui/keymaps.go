package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
)

type keyMap struct {
	// Global Key Maps
	quit         key.Binding
	toggleHelp   key.Binding
	openBoxModal key.Binding

	listPanel   listPanelKeyMap
	previewer   previewerKeyMap
	typingModal modalKeyMap
	warnModal   modalKeyMap
	fuzzyModal  fuzzyModalKeyMap
	renameInput modalKeyMap
	boxModal    boxModalKeyMap
}

type boxModalKeyMap struct {
	confirm         key.Binding
	cancel          key.Binding
	up              key.Binding
	down            key.Binding
	newBox          key.Binding
	openFolderAsBox key.Binding
	deleteBox       key.Binding
	renameBox       key.Binding
}

type listPanelKeyMap struct {
	up           key.Binding
	down         key.Binding
	newNote      key.Binding
	openTab      key.Binding
	focusPreview key.Binding
	renameNote   key.Binding
	deleteNote   key.Binding
	editNote     key.Binding
	search       key.Binding
}

type previewerKeyMap struct {
	focusList    key.Binding
	editNote     key.Binding
	up           key.Binding
	down         key.Binding
	openTab      key.Binding
	closeTab     key.Binding
	nextTab      key.Binding
	prevTab      key.Binding
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

const (
	selectionModalConfirmKey = "enter"
	selectionModalCancelKey  = "esc"
)

func defaultKeyMap() keyMap {
	vpKeys := viewport.DefaultKeyMap()

	return keyMap{
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		toggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less help"),
		),
		openBoxModal: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "boxes"),
		),
		listPanel: listPanelKeyMap{
			up: key.NewBinding(
				key.WithKeys("up", "k", "ctrl+p"),
				key.WithHelp("↑/k", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("down", "j", "ctrl+n"),
				key.WithHelp("↓/j", "down"),
			),
			newNote: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new"),
			),
			openTab: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "open tab"),
			),
			focusPreview: key.NewBinding(
				key.WithKeys("right", "l"),
				key.WithHelp("→/l", "preview"),
			),
			renameNote: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "rename"),
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
				key.WithKeys("left", "h"),
				key.WithHelp("←/h", "list"),
			),
			editNote: key.NewBinding(
				key.WithKeys("e"),
				key.WithHelp("e", "edit"),
			),
			up:   vpKeys.Up,
			down: vpKeys.Down,
			openTab: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "open tab"),
			),
			closeTab: key.NewBinding(
				key.WithKeys("w"),
				key.WithHelp("w", "close tab"),
			),
			nextTab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next tab"),
			),
			prevTab: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev tab"),
			),
			halfPageUp:   vpKeys.HalfPageUp,
			halfPageDown: vpKeys.HalfPageDown,
		},
		typingModal: modalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys(selectionModalConfirmKey),
			),
			cancel: key.NewBinding(
				key.WithKeys(selectionModalCancelKey),
			),
		},
		warnModal: modalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys(selectionModalConfirmKey),
			),
			cancel: key.NewBinding(
				key.WithKeys("q", selectionModalCancelKey),
			),
		},
		fuzzyModal: fuzzyModalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys(selectionModalConfirmKey),
			),
			cancel: key.NewBinding(
				key.WithKeys(selectionModalCancelKey),
			),
			up: key.NewBinding(
				key.WithKeys("ctrl+p", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("ctrl+n", "down"),
			),
		},
		renameInput: modalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys(selectionModalConfirmKey),
				key.WithHelp("enter", "confirm rename"),
			),
			cancel: key.NewBinding(
				key.WithKeys(selectionModalCancelKey),
				key.WithHelp("esc", "cancel"),
			),
		},
		boxModal: boxModalKeyMap{
			confirm: key.NewBinding(
				key.WithKeys(selectionModalConfirmKey),
				key.WithHelp("enter", "select"),
			),
			cancel: key.NewBinding(
				key.WithKeys(selectionModalCancelKey),
				key.WithHelp("esc", "cancel"),
			),
			up: key.NewBinding(
				key.WithKeys("ctrl+p", "up"),
			),
			down: key.NewBinding(
				key.WithKeys("ctrl+n", "down"),
			),
			newBox: key.NewBinding(
				key.WithKeys("n"),
				key.WithHelp("n", "new"),
			),
			openFolderAsBox: key.NewBinding(
				key.WithKeys("o"),
				key.WithHelp("o", "open folder"),
			),
			deleteBox: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
			renameBox: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "rename"),
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
			k.keys.listPanel.openTab,
			k.keys.listPanel.renameNote,
			k.keys.listPanel.deleteNote,
			k.keys.listPanel.search,
			k.keys.openBoxModal,
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
	case onRenaming:
		return []key.Binding{
			k.keys.renameInput.confirm,
			k.keys.renameInput.cancel,
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
			{k.keys.listPanel.newNote, k.keys.listPanel.openTab, k.keys.listPanel.renameNote, k.keys.listPanel.deleteNote, k.keys.listPanel.editNote, k.keys.listPanel.search},
			{k.keys.toggleHelp, k.keys.quit},
		}
	case onPreviewer:
		return [][]key.Binding{
			{k.keys.previewer.up, k.keys.previewer.down, k.keys.previewer.openTab},
			{k.keys.previewer.halfPageUp, k.keys.previewer.halfPageDown},
			{k.keys.previewer.focusList, k.keys.previewer.editNote},
			{k.keys.toggleHelp, k.keys.quit},
		}
	case onRenaming:
		return [][]key.Binding{
			{k.keys.renameInput.confirm, k.keys.renameInput.cancel},
		}
	default:
		return [][]key.Binding{{k.keys.toggleHelp, k.keys.quit}}
	}
}
