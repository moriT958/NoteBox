package tui

import (
	"context"
	"fmt"
	"os"

	"notebox/internal/config"
	"notebox/internal/note"
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
	previewer previewer

	// modal fields
	modalWidth  int
	modalHeight int

	// warn modal fields
	warnMessage string
	warnAction  warnAction

	// typing modal fields
	input textinput.Model

	// listpanel fields
	listPanel listPanel

	// fnsModal modal fields
	fnsModal filenameSearchModal

	// boxModal fields
	boxModal boxModal

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

	prev, err := newPreviewer(cfg)
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
		listPanel: listPanel{
			registerer:   reg,
			notesUpdates: ch,
			renameInput:  textinput.New(),
		},
		previewer: *prev,
		input:     textinput.New(),
		fnsModal: filenameSearchModal{
			input: textinput.New(),
		},
		boxModal: boxModal{
			titleInput:  textinput.New(),
			pathInput:   textinput.New(),
			renameInput: textinput.New(),
		},
		keys: defaultKeyMap(),
		help: help.New(),
	}
	// These are static values that don't change after initialization;
	// only SetWidth is called on resize in updateBoxCreateModalSize.
	m.boxModal.titleInput.Placeholder = "Box name..."
	m.boxModal.titleInput.CharLimit = 100
	m.boxModal.pathInput.Placeholder = "Path (e.g. /path/to/dir)..."
	m.boxModal.pathInput.CharLimit = 200
	m.boxModal.renameInput.CharLimit = 100
	return m, nil
}

func newBox(br note.BoxRepository) (note.Box, error) {
	ctx := context.Background()
	boxes, err := br.FindAll(ctx)
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
	return waitNoteChangeCmd(m.listPanel.notesUpdates)
}

func (m *model) handleKeyMsg(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd

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
		switch {
		case key.Matches(msg, m.keys.listPanel.down):
			m.listPanel.cursorDown()
			return m.previewer.previewNote(m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.listPanel.up):
			m.listPanel.cursorUp()
			return m.previewer.previewNote(m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.listPanel.newNote):
			m.toggleTypingModal(open)
		case key.Matches(msg, m.keys.listPanel.openTab):
			m.focus = onPreviewer
			return openNormalTabCmd(m.previewer.renderer, m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.listPanel.focusPreview):
			m.focus = onPreviewer
		case key.Matches(msg, m.keys.listPanel.deleteNote):
			m.warnMessage = "Are you sure you want to remove?"
			m.warnAction = warnDeleteNote
			m.toggleWarnModal(open)
		case key.Matches(msg, m.keys.listPanel.editNote):
			cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.selectedItem().Path)
		case key.Matches(msg, m.keys.listPanel.search):
			m.toggleFuzzyModal(open)
		case key.Matches(msg, m.keys.listPanel.renameNote):
			if m.listPanel.selectedItem().Path != "" {
				m.listPanel.renameInput.Reset()
				m.listPanel.renameInput.SetValue(m.listPanel.selectedItem().Title)
				m.listPanel.renameInput.Focus()
				m.focus = onRenaming
			}
		}
	case onRenaming:
		switch {
		case key.Matches(msg, m.keys.renameInput.confirm):
			newTitle := m.listPanel.renameInput.Value()
			m.listPanel.renameInput.Blur()
			m.focus = onListPanel
			cmd = renameNoteCmd(m.listPanel.selectedItem(), newTitle)
		case key.Matches(msg, m.keys.renameInput.cancel):
			m.listPanel.renameInput.Blur()
			m.focus = onListPanel
		default:
			m.listPanel.renameInput, cmd = m.listPanel.renameInput.Update(msg)
		}
	case onTypingModal:
		switch {
		case key.Matches(msg, m.keys.typingModal.confirm):
			m.toggleTypingModal(shut)
			cmd = createNewNoteCmd(m.currentBox.Path, m.input.Value())
		case key.Matches(msg, m.keys.typingModal.cancel):
			m.toggleTypingModal(shut)
		default:
			m.input, cmd = m.input.Update(msg)
		}
	case onPreviewer:
		switch {
		case key.Matches(msg, m.keys.previewer.focusList):
			m.focus = onListPanel
		case key.Matches(msg, m.keys.previewer.editNote):
			cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.selectedItem().Path)
		case key.Matches(msg, m.keys.previewer.openTab):
			return openNormalTabCmd(m.previewer.renderer, m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.previewer.closeTab):
			m.previewer.closeTab()
		case key.Matches(msg, m.keys.previewer.nextTab):
			m.previewer.nextTab()
		case key.Matches(msg, m.keys.previewer.prevTab):
			m.previewer.prevTab()
		default:
			m.previewer.vp, cmd = m.previewer.vp.Update(msg)
		}
	case onWarnModal:
		switch {
		case key.Matches(msg, m.keys.warnModal.confirm):
			switch m.warnAction {
			case warnDeleteNote:
				var cmds []tea.Cmd
				m.focus = onListPanel
				// WARN:
				// Do not change the execution order of deleteNotefileCmd, removeItem and renderPreviewCmd.
				// Because the cursor value is modified within removeItem, and altering
				// the order may lead to unexpected behavior.
				deletedPath := m.listPanel.selectedItem().Path
				cmds = append(cmds, deleteNoteFileCmd(deletedPath))
				m.listPanel.removeItem()
				m.previewer.removeTabByPath(deletedPath)
				cmds = append(cmds, renderPreviewCmd(m.previewer.renderer, m.listPanel.selectedItem()))
				cmd = tea.Batch(cmds...)
			case warnDeleteBox:
				selected := m.boxModal.selectedItem()
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
	case onFuzzyModal:
		switch {
		case key.Matches(msg, m.keys.fuzzyModal.confirm):
			m.selectFromFuzzy()
			m.toggleFuzzyModal(shut)
			return m.previewer.previewNote(m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.fuzzyModal.cancel):
			m.toggleFuzzyModal(shut)
		case key.Matches(msg, m.keys.fuzzyModal.down):
			m.fnsModal.cursorDown()
		case key.Matches(msg, m.keys.fuzzyModal.up):
			m.fnsModal.cursorUp()
		default:
			m.fnsModal.input, cmd = m.fnsModal.input.Update(msg)
			m.fnsModal.filter(m.fnsModal.input.Value())
		}
	case onBoxModal:
		switch {
		case key.Matches(msg, m.keys.boxModal.confirm):
			selected := m.boxModal.selectedItem()
			m.toggleBoxModal(shut)
			if selected.ID != 0 && selected.ID != m.currentBox.ID {
				return m.switchBox(selected)
			}
		case key.Matches(msg, m.keys.boxModal.cancel):
			m.toggleBoxModal(shut)
		case key.Matches(msg, m.keys.boxModal.down):
			m.boxModal.cursorDown()
		case key.Matches(msg, m.keys.boxModal.up):
			m.boxModal.cursorUp()
		case key.Matches(msg, m.keys.boxModal.newBox):
			m.toggleBoxFormModal(open, modeNewBox)
		case key.Matches(msg, m.keys.boxModal.openFolderAsBox):
			m.toggleBoxFormModal(open, modeOpenFolder)
		case key.Matches(msg, m.keys.boxModal.deleteBox):
			selected := m.boxModal.selectedItem()
			if selected.ID != 0 && selected.ID != m.currentBox.ID {
				m.warnMessage = fmt.Sprintf("Delete box '%s'?", selected.Title)
				m.warnAction = warnDeleteBox
				m.focus = onWarnModal
			}
		case key.Matches(msg, m.keys.boxModal.renameBox):
			selected := m.boxModal.selectedItem()
			if selected.ID != 0 {
				m.boxModal.renameInput.Reset()
				m.boxModal.renameInput.SetValue(selected.Title)
				m.boxModal.renameInput.Focus()
				m.focus = onBoxRenaming
			}
		}
	case onBoxRenaming:
		switch {
		case key.Matches(msg, m.keys.renameInput.confirm):
			newTitle := m.boxModal.renameInput.Value()
			m.boxModal.renameInput.Blur()
			m.focus = onBoxModal
			selected := m.boxModal.selectedItem()
			if newTitle != "" && newTitle != selected.Title {
				cmd = renameBoxCmd(m.boxRepo, note.Box{ID: selected.ID, Title: newTitle, Path: selected.Path})
			}
		case key.Matches(msg, m.keys.renameInput.cancel):
			m.boxModal.renameInput.Blur()
			m.focus = onBoxModal
		default:
			m.boxModal.renameInput, cmd = m.boxModal.renameInput.Update(msg)
		}
	case onBoxCreateModal:
		switch {
		case key.Matches(msg, m.keys.typingModal.confirm):
			cmd = m.handleBoxFormConfirm()
		case key.Matches(msg, m.keys.typingModal.cancel):
			m.toggleBoxFormModal(shut, m.boxModal.mode)
		case key.Matches(msg, m.keys.boxModal.down):
			m.boxModal.titleInput.Blur()
			m.boxModal.pathInput.Focus()
			m.boxModal.activeField = pathField
		case key.Matches(msg, m.keys.boxModal.up):
			m.boxModal.pathInput.Blur()
			m.boxModal.titleInput.Focus()
			m.boxModal.activeField = titleField
		default:
			if m.boxModal.activeField == titleField {
				m.boxModal.titleInput, cmd = m.boxModal.titleInput.Update(msg)
			} else {
				m.boxModal.pathInput, cmd = m.boxModal.pathInput.Update(msg)
			}
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
		cmd = renderPreviewCmd(m.previewer.renderer, m.listPanel.selectedItem())
	case tea.KeyPressMsg:
		cmd = m.handleKeyMsg(msg)
	case renderPreviewMsg:
		m.updatePreviewerContent(msg)
	case newNoteCreatedMsg:
		m.listPanel.addItem(note.Note(msg))
		cmd = renderPreviewCmd(m.previewer.renderer, m.listPanel.selectedItem())
	case openNormalTabMsg:
		m.previewer.openTab(msg)
	case boxesLoadedMsg:
		boxes := []note.Box(msg)
		m.boxModal.items = boxes
		m.boxModal.cursor = 0
		m.boxModal.offset = 0
		for i, b := range boxes {
			if b.ID == m.currentBox.ID {
				m.boxModal.cursor = i
				break
			}
		}
		m.focus = onBoxModal
	case boxCreatedMsg:
		m.boxModal.items = append(m.boxModal.items, note.Box(msg))
		m.focus = onBoxModal
	case boxDeletedMsg:
		deletedID := int(msg)
		for i, b := range m.boxModal.items {
			if b.ID == deletedID {
				m.boxModal.items = append(m.boxModal.items[:i], m.boxModal.items[i+1:]...)
				if m.boxModal.cursor >= len(m.boxModal.items) && m.boxModal.cursor > 0 {
					m.boxModal.cursor--
				}
				break
			}
		}
	case boxRenamedMsg:
		updated := note.Box(msg)
		for i, b := range m.boxModal.items {
			if b.ID == updated.ID {
				m.boxModal.items[i] = updated
				break
			}
		}
		if m.currentBox.ID == updated.ID {
			m.currentBox.Title = updated.Title
		}
	case notesChangedMsg:
		m.reloadAllNotes([]note.Note(msg))
		if m.focus == onFuzzyModal {
			m.fnsModal.allItems = m.listPanel.items
			m.fnsModal.filter(m.fnsModal.input.Value())
		}
		cmd = tea.Batch(waitNoteChangeCmd(m.listPanel.notesUpdates), renderPreviewCmd(m.previewer.renderer, m.listPanel.selectedItem()))
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
	return view
}

func (m model) renderOverlay(modal string, x, y int) string {
	background := lipgloss.JoinVertical(lipgloss.Center,
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
