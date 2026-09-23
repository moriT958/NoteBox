package tui

import (
	"context"
	"errors"
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"notebox/internal/config"
	"notebox/internal/core/box"
	"notebox/internal/core/note"
	"notebox/internal/tui/common"
	"notebox/internal/tui/dialog"
	"notebox/internal/tui/notelist"
	"notebox/internal/tui/preview"
	"notebox/internal/tui/styles"
	"notebox/internal/watcher"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type focusState int

const (
	focusList focusState = iota
	focusPreview
)

type UI struct {
	com *common.Common

	width, height int
	layout        uiLayout
	focus         focusState
	fullHelp      bool

	box     box.Box
	notes   *note.NoteService
	boxes   *box.BoxService
	watcher *watcher.Watcher
	// watchCh is the current box's change channel; changes reported on an
	// older one are ignored.
	watchCh  <-chan struct{}
	renderer *preview.Renderer

	list    *notelist.NoteList
	preview *preview.Preview
	status  *Status
	dialog  *dialog.Overlay
	keys    KeyMap
}

var _ tea.Model = (*UI)(nil)

// New opens the last used box (or a fallback) and starts watching it.
func New(cfg *config.Config, boxes *box.BoxService, w *watcher.Watcher) (*UI, error) {
	sty, err := styles.New(cfg.Theme)
	if err != nil {
		return nil, err
	}
	renderer, err := preview.NewRenderer(cfg.Theme)
	if err != nil {
		return nil, err
	}
	b, err := startupBox(context.Background(), boxes)
	if err != nil {
		return nil, fmt.Errorf("failed to open a box: %w", err)
	}
	ch, err := w.Watch(b.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to watch box: %w", err)
	}

	com := &common.Common{Config: cfg, Styles: sty}
	return &UI{
		com:      com,
		box:      b,
		notes:    note.NewNoteService(b),
		boxes:    boxes,
		watcher:  w,
		watchCh:  ch,
		renderer: renderer,
		list:     notelist.New(com),
		preview:  preview.New(com),
		status:   NewStatus(com),
		dialog:   dialog.NewOverlay(),
		keys:     DefaultKeyMap(),
	}, nil
}

// startupBox opens the last used box. Only when it's gone does it look
// through all boxes for one whose directory exists, and finally falls back
// to the default box.
func startupBox(ctx context.Context, svc *box.BoxService) (box.Box, error) {
	if b := lastUsedBox(ctx, svc); b != nil {
		return *b, nil
	}

	boxes, err := svc.GetActiveBoxes(ctx)
	if err != nil {
		return box.Box{}, err
	}
	for _, b := range boxes {
		if dirExists(b.Path) {
			return b, nil
		}
	}

	dir, err := config.DefaultNotesDir()
	if err != nil {
		return box.Box{}, err
	}
	for _, b := range boxes {
		if b.Path == dir {
			return b, nil
		}
	}
	b, err := svc.OpenFolderAsBox(ctx, "Default", dir)
	if err != nil {
		return box.Box{}, err
	}
	return *b, nil
}

// lastUsedBox returns the last used box if it can still be opened.
func lastUsedBox(ctx context.Context, svc *box.BoxService) *box.Box {
	id, err := config.LoadLastBoxID()
	if err != nil {
		slog.Error("failed to load last used box id", "error", err)
		return nil
	}
	if id == "" {
		return nil
	}
	b, err := svc.GetBox(ctx, id)
	if err != nil {
		slog.Error("failed to get last used box", "error", err)
		return nil
	}
	if b == nil || !b.Active || !dirExists(b.Path) {
		return nil
	}
	return b
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (m *UI) Init() tea.Cmd {
	return watchNotesCmd(m.watchCh)
}

func (m *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		return m, m.handleKey(msg)
	case tea.PasteMsg:
		if m.dialog.HasDialogs() {
			return m, m.handleAction(m.dialog.HandleMsg(msg))
		}
		if m.list.Renaming() {
			return m, m.list.UpdateRenameInput(msg)
		}

	case errMsg:
		slog.Error("tui", "error", msg.err)
		return m, m.status.SetError(msg.err)
	case clearStatusMsg:
		m.status.Clear(msg.seq)

	case notesChangedMsg:
		if msg.ch != m.watchCh {
			return m, nil
		}
		return m, tea.Batch(loadNotesCmd(m.notes, m.box.ID), watchNotesCmd(m.watchCh))
	case notesLoadedMsg:
		if msg.boxID == m.box.ID {
			return m, m.refreshNotes(msg.notes)
		}
	case noteRenderedMsg:
		if msg.boxID != m.box.ID {
			return m, nil
		}
		// A preview render for a note the cursor already left is stale.
		selected, _ := m.list.Selected()
		if msg.pin || m.preview.Has(msg.note) || msg.note == selected {
			m.preview.SetRendered(msg.note, msg.rendered, msg.pin)
		}
	case noteCreatedMsg:
		if msg.boxID == m.box.ID {
			m.list.Add(msg.note)
			return m, m.showSelected()
		}
	case noteRenamedMsg:
		if msg.boxID == m.box.ID {
			m.list.Replace(msg.old, msg.new)
			m.preview.Rename(msg.old, msg.new)
		}
	case noteRemovedMsg:
		if msg.boxID == m.box.ID {
			m.list.Remove(msg.note)
			m.preview.Remove(msg.note)
			return m, m.showSelected()
		}

	case boxesLoadedMsg:
		if d, ok := m.dialog.Find(dialog.BoxesID).(*dialog.Boxes); ok {
			d.SetBoxes(msg.items, msg.selectID)
		}
	case boxCreatedMsg:
		m.dialog.Close(dialog.BoxFormID)
		return m, loadBoxesCmd(m.boxes, msg.box.ID)
	case boxFormErrMsg:
		if d, ok := m.dialog.Find(dialog.BoxFormID).(*dialog.BoxForm); ok {
			d.SetError(msg.err)
		}
	case boxRenamedMsg:
		if msg.box.ID == m.box.ID {
			m.box = msg.box
			m.notes = note.NewNoteService(msg.box)
		}
		return m, loadBoxesCmd(m.boxes, msg.box.ID)
	case boxRemovedMsg:
		return m, loadBoxesCmd(m.boxes, "")
	}
	return m, nil
}

func (m *UI) resize(width, height int) {
	m.width, m.height = width, height
	m.layout = generateLayout(m.com.Styles, width, height)
	m.list.SetSize(m.layout.list.Dx(), m.layout.list.Dy())
	m.preview.SetSize(m.layout.preview.Dx(), m.layout.preview.Dy())
	m.status.SetWidth(width)
}

// handleKey routes a key press: quit always works, then an open dialog
// takes every key, then the inline rename, then the focused pane.
func (m *UI) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	if key.Matches(msg, m.keys.Quit) {
		return tea.Quit
	}

	if front := m.dialog.Front(); front != nil {
		if key.Matches(msg, m.keys.Boxes) && front.ID() == dialog.BoxesID {
			m.dialog.CloseFront()
			return nil
		}
		return m.handleAction(front.HandleMsg(msg))
	}

	if m.list.Renaming() {
		return m.handleRenameKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Help):
		m.fullHelp = !m.fullHelp
		return nil
	case key.Matches(msg, m.keys.Boxes):
		m.dialog.Open(dialog.NewBoxes(m.com, m.box.ID))
		return loadBoxesCmd(m.boxes, m.box.ID)
	}

	switch m.focus {
	case focusList:
		return m.handleListKey(msg)
	case focusPreview:
		return m.handlePreviewKey(msg)
	}
	return nil
}

func (m *UI) handleListKey(msg tea.KeyPressMsg) tea.Cmd {
	k := m.keys.List
	switch {
	case key.Matches(msg, k.Up):
		m.list.CursorUp()
		return m.showSelected()
	case key.Matches(msg, k.Down):
		m.list.CursorDown()
		return m.showSelected()
	case key.Matches(msg, k.FocusPreview):
		m.focus = focusPreview
		return nil
	case key.Matches(msg, k.New):
		m.dialog.Open(dialog.NewNoteInput(m.com))
		return nil
	case key.Matches(msg, k.Search):
		m.dialog.Open(dialog.NewFinder(m.com, m.list.Notes()))
		return nil
	}

	n, ok := m.list.Selected()
	if !ok {
		return nil
	}
	switch {
	case key.Matches(msg, k.Open):
		m.focus = focusPreview
		return m.open(n, true)
	case key.Matches(msg, k.Rename):
		m.list.StartRename()
	case key.Matches(msg, k.Delete):
		message := fmt.Sprintf("Delete note '%s'?", n.Title())
		m.dialog.Open(dialog.NewConfirm(m.com, message, dialog.ActionDeleteNote{Note: n}))
	case key.Matches(msg, k.Edit):
		return m.edit(n)
	}
	return nil
}

func (m *UI) handleRenameKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Rename.Cancel):
		m.list.StopRename()
	case key.Matches(msg, m.keys.Rename.Confirm):
		title := m.list.RenameValue()
		m.list.StopRename()
		n, ok := m.list.Selected()
		if !ok || title == "" || title == n.Title() {
			return nil
		}
		return renameNoteCmd(m.notes, m.box.ID, n, title)
	default:
		return m.list.UpdateRenameInput(msg)
	}
	return nil
}

func (m *UI) handlePreviewKey(msg tea.KeyPressMsg) tea.Cmd {
	k := m.keys.Preview
	switch {
	case key.Matches(msg, k.FocusList):
		m.focus = focusList
	case key.Matches(msg, k.Edit):
		if n, ok := m.preview.Active(); ok {
			return m.edit(n)
		}
	case key.Matches(msg, k.Pin):
		m.preview.PinActive()
	case key.Matches(msg, k.CloseTab):
		m.preview.CloseActive()
	case key.Matches(msg, k.NextTab):
		m.preview.NextTab()
	case key.Matches(msg, k.PrevTab):
		m.preview.PrevTab()
	default:
		return m.preview.Scroll(msg)
	}
	return nil
}

// handleAction carries out what a dialog asked for.
func (m *UI) handleAction(action dialog.Action) tea.Cmd {
	switch a := action.(type) {
	case dialog.ActionClose:
		m.dialog.CloseFront()
	case dialog.ActionOpen:
		m.dialog.Open(a.Dialog)
	case dialog.ActionCmd:
		return a.Cmd

	case dialog.ActionCreateNote:
		m.dialog.CloseFront()
		return createNoteCmd(m.notes, m.box.ID, a.Title)
	case dialog.ActionDeleteNote:
		m.dialog.CloseFront()
		return removeNoteCmd(m.notes, m.box.ID, a.Note)
	case dialog.ActionSelectNote:
		m.dialog.CloseFront()
		m.list.Select(a.Note)
		m.focus = focusList
		return m.showSelected()

	case dialog.ActionSwitchBox:
		m.dialog.Close(dialog.BoxesID)
		if a.Box.ID == m.box.ID {
			return nil
		}
		return m.switchBox(a.Box)
	case dialog.ActionCreateBox:
		// The form stays open until the box is created, to show errors.
		return createBoxCmd(m.boxes, a.Title, a.Path)
	case dialog.ActionOpenFolder:
		return openFolderCmd(m.boxes, a.Title, a.Path)
	case dialog.ActionRenameBox:
		return renameBoxCmd(m.boxes, a.Box.ID, a.Title)
	case dialog.ActionDeleteBox:
		m.dialog.CloseFront()
		return removeBoxCmd(m.boxes, a.Box.ID)
	}
	return nil
}

func (m *UI) switchBox(b box.Box) tea.Cmd {
	ch, err := m.watcher.Watch(b.Path)
	if err != nil {
		return m.status.SetError(fmt.Errorf("failed to open box: %w", err))
	}
	m.box = b
	m.notes = note.NewNoteService(b)
	m.watchCh = ch
	m.list.StopRename()
	m.list.SetNotes(nil)
	m.preview.Clear()
	m.focus = focusList
	return tea.Batch(watchNotesCmd(ch), saveLastBoxCmd(b.ID))
}

// refreshNotes updates the list and the preview with a freshly loaded note
// list: tabs of notes that are gone are closed, the rest are re-rendered
// since their files may have changed (e.g. edited in $EDITOR), and when no
// tab is open (right after startup or a box switch) the selected note is
// previewed.
func (m *UI) refreshNotes(notes []note.Note) tea.Cmd {
	m.list.SetNotes(notes)

	var cmds []tea.Cmd
	for _, n := range m.preview.Notes() {
		if !slices.Contains(notes, n) {
			m.preview.Remove(n)
			continue
		}
		cmds = append(cmds, renderNoteCmd(m.notes, m.renderer, m.box.ID, n, false))
	}
	if len(m.preview.Notes()) == 0 {
		cmds = append(cmds, m.showSelected())
	}
	return tea.Batch(cmds...)
}

// showSelected previews the note under the list cursor.
func (m *UI) showSelected() tea.Cmd {
	n, ok := m.list.Selected()
	if !ok {
		return nil
	}
	return m.open(n, false)
}

// open shows n in the preview, rendering it first if it has no tab yet.
func (m *UI) open(n note.Note, pin bool) tea.Cmd {
	if m.preview.Open(n, pin) {
		return nil
	}
	return renderNoteCmd(m.notes, m.renderer, m.box.ID, n, pin)
}

func (m *UI) edit(n note.Note) tea.Cmd {
	editor := m.com.Config.Editor
	if editor == "" {
		return m.status.SetError(errors.New("no editor configured"))
	}
	return editNoteCmd(editor, filepath.Join(m.box.Path, n.Path()))
}

// helpKeys returns the keys that currently apply, for the help line.
func (m *UI) helpKeys() help.KeyMap {
	switch {
	case m.dialog.HasDialogs():
		return m.dialog.Front()
	case m.list.Renaming():
		return m.keys.renameHelp()
	case m.focus == focusPreview:
		return m.keys.previewHelp(m.preview.ViewportKeyMap())
	default:
		return m.keys.listHelp()
	}
}

func (m *UI) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "Note Box"
	if m.width == 0 || m.height == 0 {
		return v
	}

	l := m.layout
	keys := m.helpKeys()
	header := m.com.Styles.Header.Width(l.header.Dx()).Render(m.box.Title)
	layers := []*lipgloss.Layer{
		place(header, l.header.Min, 0),
		place(m.list.Render(m.focus == focusList), l.list.Min, 0),
		place(m.preview.Render(m.focus == focusPreview), l.preview.Min, 0),
		place(m.status.Render(l.status.Dx(), keys), l.status.Min, 0),
	}
	if c := m.list.Cursor(); c != nil {
		v.Cursor = offsetCursor(c, l.list.Min)
	}

	if m.fullHelp && !m.dialog.HasDialogs() {
		full := m.status.RenderFullHelp(m.width, keys)
		layers = append(layers, place(full, image.Pt(0, max(0, m.height-lipgloss.Height(full))), 1))
	}

	for i, d := range m.dialog.Dialogs() {
		content, cursor := d.Render(l.area)
		rect := common.CenterRect(l.area, lipgloss.Width(content), lipgloss.Height(content))
		layers = append(layers, place(content, rect.Min, 2+i))
		// Dialogs are drawn back to front; the front one owns the cursor.
		v.Cursor = nil
		if cursor != nil {
			v.Cursor = offsetCursor(cursor, rect.Min)
		}
	}

	canvas := lipgloss.NewCanvas(m.width, m.height)
	canvas.Compose(lipgloss.NewCompositor(layers...))
	v.Content = canvas.Render()
	return v
}

func place(content string, at image.Point, z int) *lipgloss.Layer {
	return lipgloss.NewLayer(content).X(at.X).Y(at.Y).Z(z)
}

func offsetCursor(c *tea.Cursor, by image.Point) *tea.Cursor {
	c.Position.X += by.X
	c.Position.Y += by.Y
	return c
}
