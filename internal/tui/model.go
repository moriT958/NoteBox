package tui

import (
	"context"
	"fmt"
	"os"

	"notebox/internal/config"
	"notebox/internal/note"
	"notebox/internal/tui/boxmodal"
	"notebox/internal/tui/fuzzymodal"
	"notebox/internal/tui/listpanel"
	"notebox/internal/tui/previewer"
	"notebox/internal/tui/styles"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type focus int

const (
	onListPanel focus = iota
	onPreviewer
	onTypingModal
	onWarnModal
	onFuzzyModal
	onRenaming
	onBoxModal
	onBoxCreateModal
	onBoxRenaming
)

type warnAction int

const (
	warnDeleteNote warnAction = iota
	warnDeleteBox
)

const (
	helpGuideHeight = 1
	headerHeight    = 2 // header text line + bottom border
)

type model struct {
	cfg        *config.Config
	styles     *styles.Style
	currentBox note.Box
	boxRepo    note.BoxRepository

	// main model fields
	width, height int
	focus         focus

	// previewer fields
	previewer previewer.Previewer

	// modal fields
	modalWidth  int
	modalHeight int

	// warn modal fields
	warnMessage string
	warnAction  warnAction

	// typing modal fields
	input textinput.Model

	// listpanel fields
	listPanel listpanel.ListPanel

	// fnsModal modal fields
	fnsModal fuzzymodal.Modal

	// boxModal fields
	boxModal boxmodal.BoxModal

	// key / help fields
	keys keyMap
	help help.Model
}

func NewModel(reg note.Registerer, br note.BoxRepository) (*model, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	theme, err := styles.GetColorTheme(cfg)
	if err != nil {
		return nil, err
	}

	curBox, err := newBox(br)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize box: %w", err)
	}

	ch, err := reg.Register(curBox.Path)
	if err != nil {
		return nil, err
	}

	prev, err := previewer.New(cfg)
	if err != nil {
		return nil, err
	}

	m := &model{
		cfg:         cfg,
		styles:      styles.New(theme),
		currentBox:  curBox,
		boxRepo:     br,
		width:       0,
		height:      0,
		modalWidth:  60,
		modalHeight: 7,
		focus:       onListPanel,
		listPanel: listpanel.ListPanel{
			Registerer:   reg,
			NotesUpdates: ch,
			RenameInput:  textinput.New(),
		},
		previewer: *prev,
		input:     textinput.New(),
		fnsModal: fuzzymodal.Modal{
			Input: textinput.New(),
		},
		boxModal: boxmodal.BoxModal{
			TitleInput:  textinput.New(),
			PathInput:   textinput.New(),
			RenameInput: textinput.New(),
		},
		keys: defaultKeyMap(),
		help: help.New(),
	}
	// These are static values that don't change after initialization;
	// only SetWidth is called on resize in updateBoxCreateModalSize.
	m.boxModal.TitleInput.Placeholder = "Box name..."
	m.boxModal.TitleInput.CharLimit = 100
	m.boxModal.PathInput.Placeholder = "Path (e.g. /path/to/dir)..."
	m.boxModal.PathInput.CharLimit = 200
	m.boxModal.RenameInput.CharLimit = 100

	m.input.SetVirtualCursor(false)
	m.listPanel.RenameInput.SetVirtualCursor(false)
	m.fnsModal.Input.SetVirtualCursor(false)
	m.boxModal.TitleInput.SetVirtualCursor(false)
	m.boxModal.PathInput.SetVirtualCursor(false)
	m.boxModal.RenameInput.SetVirtualCursor(false)

	return m, nil
}

func newBox(br note.BoxRepository) (note.Box, error) {
	ctx := context.Background()
	boxes, err := br.FindAllActive(ctx)
	if err != nil {
		return note.Box{}, err
	}

	if len(boxes) == 0 {
		defaultPath, err := config.DefaultNotesDir()
		if err != nil {
			return note.Box{}, err
		}
		return br.CreateBox(ctx, note.Box{Title: "Default", Path: defaultPath})
	}

	if lastID, err := config.LoadLastBoxID(); err == nil && lastID > 0 {
		for _, b := range boxes {
			if b.ID == lastID && pathExists(b.Path) {
				return b, nil
			}
		}
	}

	for _, b := range boxes {
		if pathExists(b.Path) {
			return b, nil
		}
	}

	defaultPath, err := config.DefaultNotesDir()
	if err != nil {
		return note.Box{}, err
	}
	return br.CreateBox(ctx, note.Box{Title: "Default", Path: defaultPath})
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (m model) Init() tea.Cmd {
	return waitNoteChangeCmd(m.listPanel.NotesUpdates)
}

// handleKeyMsg handles the few keys that apply regardless of focus, then
// dispatches to the handler for the current focus state. Each handler owns
// its own key bindings and invariants and can be driven directly in tests
// without going through this dispatcher.
func (m *model) handleKeyMsg(msg tea.KeyPressMsg) tea.Cmd {
	if key.Matches(msg, m.keys.quit) {
		return tea.Quit
	}
	if key.Matches(msg, m.keys.toggleHelp) {
		m.help.ShowAll = !m.help.ShowAll
		return nil
	}
	if key.Matches(msg, m.keys.openBoxModal) {
		if m.focus == onBoxModal {
			m.toggleBoxModal(shut)
			return nil
		}
		return loadBoxesCmd(m.boxRepo)
	}

	switch m.focus {
	case onListPanel:
		return m.handleListPanelKeys(msg)
	case onRenaming:
		return m.handleRenamingKeys(msg)
	case onTypingModal:
		return m.handleTypingModalKeys(msg)
	case onPreviewer:
		return m.handlePreviewerKeys(msg)
	case onWarnModal:
		return m.handleWarnModalKeys(msg)
	case onFuzzyModal:
		return m.handleFuzzyModalKeys(msg)
	case onBoxModal:
		return m.handleBoxModalKeys(msg)
	case onBoxRenaming:
		return m.handleBoxRenamingKeys(msg)
	case onBoxCreateModal:
		return m.handleBoxCreateModalKeys(msg)
	}
	return nil
}

func (m *model) handleListPanelKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.listPanel.down):
		m.listPanel.CursorDown()
		return m.previewer.OpenTab(m.listPanel.SelectedItem(), false)
	case key.Matches(msg, m.keys.listPanel.up):
		m.listPanel.CursorUp()
		return m.previewer.OpenTab(m.listPanel.SelectedItem(), false)
	case key.Matches(msg, m.keys.listPanel.newNote):
		m.toggleTypingModal(open)
	case key.Matches(msg, m.keys.listPanel.openTab):
		m.focus = onPreviewer
		return m.previewer.OpenTab(m.listPanel.SelectedItem(), true)
	case key.Matches(msg, m.keys.listPanel.focusPreview):
		m.focus = onPreviewer
	case key.Matches(msg, m.keys.listPanel.deleteNote):
		m.warnMessage = "Are you sure you want to remove?"
		m.warnAction = warnDeleteNote
		m.toggleWarnModal(open)
	case key.Matches(msg, m.keys.listPanel.editNote):
		cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.SelectedItem().Path)
	case key.Matches(msg, m.keys.listPanel.search):
		m.toggleFuzzyModal(open)
	case key.Matches(msg, m.keys.listPanel.renameNote):
		if m.listPanel.SelectedItem().Path != "" {
			m.listPanel.RenameInput.Reset()
			m.listPanel.RenameInput.SetValue(m.listPanel.SelectedItem().Title)
			m.listPanel.RenameInput.Focus()
			m.focus = onRenaming
		}
	}
	return cmd
}

func (m *model) handleRenamingKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.renameInput.confirm):
		newTitle := m.listPanel.RenameInput.Value()
		m.listPanel.RenameInput.Blur()
		m.focus = onListPanel
		cmd = renameNoteCmd(m.listPanel.SelectedItem(), newTitle)
	case key.Matches(msg, m.keys.renameInput.cancel):
		m.listPanel.RenameInput.Blur()
		m.focus = onListPanel
	default:
		m.listPanel.RenameInput, cmd = m.listPanel.RenameInput.Update(msg)
	}
	return cmd
}

func (m *model) handleTypingModalKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.typingModal.confirm):
		m.toggleTypingModal(shut)
		cmd = createNewNoteCmd(m.currentBox.Path, m.input.Value())
	case key.Matches(msg, m.keys.typingModal.cancel):
		m.toggleTypingModal(shut)
	default:
		m.input, cmd = m.input.Update(msg)
	}
	return cmd
}

func (m *model) handlePreviewerKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.previewer.focusList):
		m.focus = onListPanel
	case key.Matches(msg, m.keys.previewer.editNote):
		cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.SelectedItem().Path)
	case key.Matches(msg, m.keys.previewer.openTab):
		return m.previewer.OpenTab(m.listPanel.SelectedItem(), true)
	case key.Matches(msg, m.keys.previewer.closeTab):
		m.previewer.CloseTab()
	case key.Matches(msg, m.keys.previewer.nextTab):
		m.previewer.NextTab()
	case key.Matches(msg, m.keys.previewer.prevTab):
		m.previewer.PrevTab()
	default:
		m.previewer.VP, cmd = m.previewer.VP.Update(msg)
	}
	return cmd
}

func (m *model) handleWarnModalKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.warnModal.confirm):
		switch m.warnAction {
		case warnDeleteNote:
			var cmds []tea.Cmd
			m.focus = onListPanel
			// WARN:
			// Do not change the execution order of deleteNoteFileCmd, removeItem and previewer.Refresh.
			// Because the cursor value is modified within removeItem, and altering
			// the order may lead to unexpected behavior.
			deletedPath := m.listPanel.SelectedItem().Path
			cmds = append(cmds, deleteNoteFileCmd(deletedPath))
			m.listPanel.RemoveItem()
			m.previewer.RemoveTabByPath(deletedPath)
			cmds = append(cmds, m.previewer.Refresh(m.listPanel.SelectedItem()))
			cmd = tea.Batch(cmds...)
		case warnDeleteBox:
			selected := m.boxModal.SelectedItem()
			m.focus = onBoxModal
			cmd = deleteBoxCmd(m.boxRepo, selected)
		}
	case key.Matches(msg, m.keys.warnModal.cancel):
		switch m.warnAction {
		case warnDeleteNote:
			m.focus = onListPanel
		case warnDeleteBox:
			m.focus = onBoxModal
		}
	}
	return cmd
}

func (m *model) handleFuzzyModalKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.fuzzyModal.confirm):
		m.selectFromFuzzy()
		m.toggleFuzzyModal(shut)
		return m.previewer.OpenTab(m.listPanel.SelectedItem(), false)
	case key.Matches(msg, m.keys.fuzzyModal.cancel):
		m.toggleFuzzyModal(shut)
	case key.Matches(msg, m.keys.fuzzyModal.down):
		m.fnsModal.CursorDown()
	case key.Matches(msg, m.keys.fuzzyModal.up):
		m.fnsModal.CursorUp()
	default:
		m.fnsModal.Input, cmd = m.fnsModal.Input.Update(msg)
		m.fnsModal.Filter(m.fnsModal.Input.Value())
	}
	return cmd
}

func (m *model) handleBoxModalKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.boxModal.confirm):
		selected := m.boxModal.SelectedItem()
		m.toggleBoxModal(shut)
		if selected.ID != 0 && selected.ID != m.currentBox.ID {
			return m.switchBox(selected)
		}
	case key.Matches(msg, m.keys.boxModal.cancel):
		m.toggleBoxModal(shut)
	case key.Matches(msg, m.keys.boxModal.down):
		m.boxModal.CursorDown()
	case key.Matches(msg, m.keys.boxModal.up):
		m.boxModal.CursorUp()
	case key.Matches(msg, m.keys.boxModal.newBox):
		m.toggleBoxFormModal(open, boxmodal.ModeNewBox)
	case key.Matches(msg, m.keys.boxModal.openFolderAsBox):
		m.toggleBoxFormModal(open, boxmodal.ModeOpenFolder)
	case key.Matches(msg, m.keys.boxModal.deleteBox):
		selected := m.boxModal.SelectedItem()
		if selected.ID != 0 && selected.ID != m.currentBox.ID {
			m.warnMessage = fmt.Sprintf("Delete box '%s'?", selected.Title)
			m.warnAction = warnDeleteBox
			m.focus = onWarnModal
		}
	case key.Matches(msg, m.keys.boxModal.renameBox):
		selected := m.boxModal.SelectedItem()
		if selected.ID != 0 {
			m.boxModal.RenameInput.Reset()
			m.boxModal.RenameInput.SetValue(selected.Title)
			m.boxModal.RenameInput.Focus()
			m.focus = onBoxRenaming
		}
	}
	return cmd
}

func (m *model) handleBoxRenamingKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.renameInput.confirm):
		newTitle := m.boxModal.RenameInput.Value()
		m.boxModal.RenameInput.Blur()
		m.focus = onBoxModal
		selected := m.boxModal.SelectedItem()
		if newTitle != "" && newTitle != selected.Title {
			cmd = renameBoxCmd(m.boxRepo, note.Box{ID: selected.ID, Title: newTitle, Path: selected.Path})
		}
	case key.Matches(msg, m.keys.renameInput.cancel):
		m.boxModal.RenameInput.Blur()
		m.focus = onBoxModal
	default:
		m.boxModal.RenameInput, cmd = m.boxModal.RenameInput.Update(msg)
	}
	return cmd
}

func (m *model) handleBoxCreateModalKeys(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keys.typingModal.confirm):
		cmd = m.handleBoxFormConfirm()
	case key.Matches(msg, m.keys.typingModal.cancel):
		m.toggleBoxFormModal(shut, m.boxModal.Mode)
	case key.Matches(msg, m.keys.boxModal.down):
		m.boxModal.TitleInput.Blur()
		m.boxModal.PathInput.Focus()
		m.boxModal.ActiveField = boxmodal.PathField
	case key.Matches(msg, m.keys.boxModal.up):
		m.boxModal.PathInput.Blur()
		m.boxModal.TitleInput.Focus()
		m.boxModal.ActiveField = boxmodal.TitleField
	default:
		if m.boxModal.ActiveField == boxmodal.TitleField {
			m.boxModal.TitleInput, cmd = m.boxModal.TitleInput.Update(msg)
		} else {
			m.boxModal.PathInput, cmd = m.boxModal.PathInput.Update(msg)
		}
	}
	return cmd
}

func (m *model) handlePasteMsg(msg tea.PasteMsg) tea.Cmd {
	var cmd tea.Cmd

	switch m.focus {
	case onTypingModal:
		m.input, cmd = m.input.Update(msg)
	case onRenaming:
		m.listPanel.RenameInput, cmd = m.listPanel.RenameInput.Update(msg)
	case onFuzzyModal:
		m.fnsModal.Input, cmd = m.fnsModal.Input.Update(msg)
		m.fnsModal.Filter(m.fnsModal.Input.Value())
	case onBoxRenaming:
		m.boxModal.RenameInput, cmd = m.boxModal.RenameInput.Update(msg)
	case onBoxCreateModal:
		if m.boxModal.ActiveField == boxmodal.TitleField {
			m.boxModal.TitleInput, cmd = m.boxModal.TitleInput.Update(msg)
		} else {
			m.boxModal.PathInput, cmd = m.boxModal.PathInput.Update(msg)
		}
	}
	return cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height, m.width = msg.Height, msg.Width
		m.updateListPanelSize(msg)
		m.updatePreviewerSize(msg)
		m.updateTypingModalSize(msg)
		m.updateFuzzyModalSize(msg)
		m.updateBoxModalSize(msg)
		m.updateBoxCreateModalSize(msg)
		m.help.SetWidth(msg.Width)
		cmd = m.previewer.Refresh(m.listPanel.SelectedItem())
	case tea.KeyPressMsg:
		cmd = m.handleKeyMsg(msg)
	case tea.PasteMsg:
		cmd = m.handlePasteMsg(msg)
	case previewer.TabRenderedMsg:
		m.previewer.ApplyRendered(msg)
	case newNoteCreatedMsg:
		m.listPanel.AddItem(note.Note(msg))
		cmd = m.previewer.Refresh(m.listPanel.SelectedItem())
	case boxesLoadedMsg:
		m.boxModal.ApplyBoxesLoaded([]note.Box(msg), m.currentBox.ID)
		m.focus = onBoxModal
	case boxCreatedMsg:
		m.boxModal.ApplyBoxCreated(note.Box(msg))
		m.focus = onBoxModal
	case boxDeletedMsg:
		m.boxModal.ApplyBoxDeleted(int(msg))
	case boxRenamedMsg:
		updated := note.Box(msg)
		m.boxModal.ApplyBoxRenamed(updated)
		if m.currentBox.ID == updated.ID {
			m.currentBox.Title = updated.Title
		}
	case notesChangedMsg:
		m.listPanel.ReloadAllNotes([]note.Note(msg))
		if m.focus == onFuzzyModal {
			m.fnsModal.AllItems = m.listPanel.Items
			m.fnsModal.Filter(m.fnsModal.Input.Value())
		}
		cmd = tea.Batch(waitNoteChangeCmd(m.listPanel.NotesUpdates), m.previewer.Refresh(m.listPanel.SelectedItem()))
	}

	return m, cmd
}

func (m model) View() tea.View {
	var content string

	switch m.focus {
	case onTypingModal:
		content = m.viewTypingModal()
	case onWarnModal:
		content = m.viewWarnModal()
	case onFuzzyModal:
		content = m.viewFuzzyModal()
	case onBoxModal, onBoxRenaming:
		content = m.viewBoxModal()
	case onBoxCreateModal:
		content = m.viewBoxCreateModal()
	default:
		content = m.styles.Main.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				m.viewHeader(),
				lipgloss.JoinHorizontal(lipgloss.Top,
					m.viewListPanel(),
					m.viewPreviewer(),
				),
				m.viewHelp(),
			))
	}

	if overlay := m.viewFullHelpOverlay(); overlay != "" {
		overlayY := max(0, m.height-lipgloss.Height(overlay))
		fgLayer := lipgloss.NewLayer(overlay).X(0).Y(overlayY).Z(1)
		bgLayer := lipgloss.NewLayer(content).X(0).Y(0).Z(0)
		compositor := lipgloss.NewCompositor(bgLayer, fgLayer)
		content = lipgloss.NewCanvas(m.width, m.height).Compose(compositor).Render()
	}

	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = "Note Box"
	view.Cursor = m.cursorForFocus()
	return view
}

func (m model) cursorForFocus() *tea.Cursor {
	switch m.focus {
	case onTypingModal:
		return m.typingModalCursor()
	case onBoxCreateModal:
		return m.boxCreateModalCursor()
	case onFuzzyModal:
		return m.fuzzyModalCursor()
	case onBoxRenaming:
		return m.boxRenameCursor()
	case onRenaming:
		return m.listRenameCursor()
	default:
		return nil
	}
}

func (m model) renderOverlay(modal string, x, y int) string {
	background := lipgloss.JoinVertical(lipgloss.Center,
		m.viewHeader(),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewListPanel(),
			m.viewPreviewer(),
		),
		m.viewHelp(),
	)
	fgLayer := lipgloss.NewLayer(modal).X(x).Y(y).Z(1)
	bgLayer := lipgloss.NewLayer(background).X(0).Y(0).Z(0)
	compositor := lipgloss.NewCompositor(bgLayer, fgLayer)
	canvas := lipgloss.NewCanvas(m.width, m.height).Compose(compositor)
	return m.styles.Main.Render(canvas.Render())
}
