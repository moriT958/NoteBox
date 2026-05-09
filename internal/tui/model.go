package tui

import (
	"notebox/internal/config"
	"notebox/internal/note"
	"notebox/internal/tui/styles"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
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
)

const (
	layoutFramePadding = 4
	helpGuideHeight    = 1
)

type model struct {
	cfg    *config.Config
	styles *styles.Style

	// main model fields
	width, height int
	focus         focus

	// previewer fields
	vp       viewport.Model
	renderer note.NoteRenderer

	// modal fields
	modalWidth  int
	modalHeight int

	// typing modal fields
	input textinput.Model

	// listpanel fields
	listPanel listPanel

	// fnsModal modal fields
	fnsModal filenameSearchModal

	// key / help fields
	keys keyMap
	help help.Model
}

func NewModel(reg note.Registerer) (*model, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	theme, err := styles.GetColorTheme(cfg)
	if err != nil {
		return nil, err
	}

	r, err := note.NewGlamourRenderer(cfg.Theme)
	if err != nil {
		return nil, err
	}

	ch, err := reg.Register(cfg.NotesDir)
	if err != nil {
		return nil, err
	}

	vp := viewport.New(
		viewport.WithWidth(0),
		viewport.WithHeight(0),
	)
	vp.SetHorizontalStep(4)

	m := &model{
		cfg:         cfg,
		styles:      styles.New(theme),
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
		vp:       vp,
		renderer: r,
		input:    textinput.New(),
		fnsModal: filenameSearchModal{
			input: textinput.New(),
		},
		keys: defaultKeyMap(),
		help: help.New(),
	}
	return m, nil
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

	switch m.focus {
	case onListPanel:
		switch {
		case key.Matches(msg, m.keys.listPanel.down):
			m.listPanel.cursorDown()
			return renderPreviewCmd(m.renderer, m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.listPanel.up):
			m.listPanel.cursorUp()
			return renderPreviewCmd(m.renderer, m.listPanel.selectedItem())
		case key.Matches(msg, m.keys.listPanel.newNote):
			m.toggleTypingModal(open)
		case key.Matches(msg, m.keys.listPanel.focusPreview):
			m.focus = onPreviewer
		case key.Matches(msg, m.keys.listPanel.deleteNote):
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
			cmd = createNewNoteCmd(m.cfg.NotesDir, m.input.Value())
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
		default:
			m.vp, cmd = m.vp.Update(msg)
		}
	case onWarnModal:
		switch {
		case key.Matches(msg, m.keys.warnModal.confirm):
			var cmds []tea.Cmd
			m.toggleWarnModal(shut)
			// WARN:
			// Do not change the execution order of deleteNotefileCmd, removeItem and renderPreviewCmd.
			// Because the cursor value is modified within removeItem, and altering
			// the order may lead to unexpected behavior.
			cmds = append(cmds, deleteNoteFileCmd(m.listPanel.selectedItem().Path))
			m.listPanel.removeItem()
			cmds = append(cmds, renderPreviewCmd(m.renderer, m.listPanel.selectedItem()))
			cmd = tea.Batch(cmds...)
		case key.Matches(msg, m.keys.warnModal.cancel):
			m.toggleWarnModal(shut)
		}
	case onFuzzyModal:
		switch {
		case key.Matches(msg, m.keys.fuzzyModal.confirm):
			m.selectFromFuzzy()
			m.toggleFuzzyModal(shut)
			return renderPreviewCmd(m.renderer, m.listPanel.selectedItem())
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
		m.help.SetWidth(msg.Width)
		cmd = renderPreviewCmd(m.renderer, m.listPanel.selectedItem())
	case tea.KeyPressMsg:
		cmd = m.handleKeyMsg(msg)
	case renderPreviewMsg:
		cmd = m.updatePreviewerContent(msg)
	case newNoteCreatedMsg:
		m.listPanel.addItem(note.Note(msg))
		cmd = renderPreviewCmd(m.renderer, m.listPanel.selectedItem())
	case notesChangedMsg:
		m.reloadAllNotes([]note.Note(msg))
		if m.focus == onFuzzyModal {
			m.fnsModal.allItems = m.listPanel.items
			m.fnsModal.filter(m.fnsModal.input.Value())
		}
		cmd = tea.Batch(waitNoteChangeCmd(m.listPanel.notesUpdates), renderPreviewCmd(m.renderer, m.listPanel.selectedItem()))
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
	return view
}

func (m model) viewHeader() string {
	return m.styles.Header.
		Align(lipgloss.Center).
		Width(m.width).
		Render("📓 NoteBox 📓")
}
