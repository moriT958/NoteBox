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
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

type focus int

const (
	onListPanel focus = iota
	onPreviewer
	onTypingModal
	onWarnModal
	onFuzzyModal
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
	renderer *glamour.TermRenderer

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
	keys     keyMap
	help     help.Model
	showHelp bool
}

func NewModel() (*model, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	theme, err := styles.GetColorTheme(cfg)
	if err != nil {
		return nil, err
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(getGlamourTheme(theme)),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return nil, err
	}

	notes, err := note.LoadNoteFiles(cfg.NotesDir)
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
			cursor: 0,
			items:  notes,
			offset: 0,
		},
		vp:       vp,
		renderer: r,
		input:    textinput.New(),
		fnsModal: filenameSearchModal{
			input: textinput.New(),
		},
		keys:     defaultKeyMap(),
		help:     help.New(),
		showHelp: true,
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) handleKeyMsg(msg tea.KeyPressMsg) tea.Cmd {
	var cmd tea.Cmd

	if key.Matches(msg, m.keys.quit) {
		return tea.Quit
	}
	if key.Matches(msg, m.keys.toggleHelp) {
		m.showHelp = !m.showHelp
		return nil
	}

	switch m.focus {
	case onListPanel:
		switch {
		case key.Matches(msg, m.keys.listPanel.down):
			m.listPanel.cursorDown()
			return m.renderPreviewCmd(m.listPanel.selectedItem().Path)
		case key.Matches(msg, m.keys.listPanel.up):
			m.listPanel.cursorUp()
			return m.renderPreviewCmd(m.listPanel.selectedItem().Path)
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
			cmds = append(cmds, m.renderPreviewCmd(m.listPanel.selectedItem().Path))
			cmd = tea.Batch(cmds...)
		case key.Matches(msg, m.keys.warnModal.cancel):
			m.toggleWarnModal(shut)
		}
	case onFuzzyModal:
		switch {
		case key.Matches(msg, m.keys.fuzzyModal.confirm):
			m.selectFromFuzzy()
			m.toggleFuzzyModal(shut)
			return m.renderPreviewCmd(m.listPanel.selectedItem().Path)
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
		cmd = m.renderPreviewCmd(m.listPanel.selectedItem().Path)
	case tea.KeyPressMsg:
		cmd = m.handleKeyMsg(msg)
	case renderPreviewMsg:
		cmd = m.updatePreviewerContent(msg)
	case newNoteCreatedMsg:
		m.listPanel.addItem(note.Note(msg))
		cmd = m.renderPreviewCmd(m.listPanel.selectedItem().Path)
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

func (m model) viewHelp() string {
	if !m.showHelp {
		return ""
	}

	guideFocus := onListPanel
	if m.focus == onPreviewer {
		guideFocus = onPreviewer
	}

	focused := m.keys.forFocus(guideFocus)
	guide := m.help.ShortHelpView(focused.ShortHelp())

	return m.styles.Help.
		Width(m.width).
		Render(guide)
}
