package tui

import (
	"notebox/internal/config"
	"notebox/internal/note"
	uistyles "notebox/internal/tui/styles"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	gstyles "charm.land/glamour/v2/styles"
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

type model struct {
	cfg    *config.Config
	styles *uistyles.Style

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
}

func NewModel() (*model, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(gstyles.DarkStyle),
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
		styles:      uistyles.New(uistyles.DarkTheme),
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
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	if msg.String() == "q" {
		return tea.Quit
	}

	switch m.focus {
	case onListPanel:
		switch msg.String() {
		case "j":
			m.listPanel.cursorDown()
			return m.renderPreviewCmd(m.listPanel.selectedItem().Path)
		case "k":
			m.listPanel.cursorUp()
			return m.renderPreviewCmd(m.listPanel.selectedItem().Path)
		case "n":
			m.toggleTypingModal(open)
		case "ctrl+l":
			m.focus = onPreviewer
		case "d":
			m.toggleWarnModal(open)
		case "e":
			cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.selectedItem().Path)
		case "/":
			m.toggleFuzzyModal(open)
		}
	case onTypingModal:
		switch msg.String() {
		case "enter":
			m.toggleTypingModal(shut)
			cmd = createNewNoteCmd(m.cfg.NotesDir, m.input.Value())
		case "ctrl+c":
			m.toggleTypingModal(shut)
		default:
			m.input, cmd = m.input.Update(msg)
		}
	case onPreviewer:
		switch msg.String() {
		case "ctrl+h":
			m.focus = onListPanel
		case "e":
			cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.selectedItem().Path)
		default:
			m.vp, cmd = m.vp.Update(msg)
		}
	case onWarnModal:
		switch msg.String() {
		case "enter":
			var cmds []tea.Cmd
			m.toggleTypingModal(shut)
			// WARN:
			// Do not change the execution order of deleteNotefileCmd, removeItem and renderPreviewCmd.
			// Because the cursor value is modified within removeItem, and altering
			// the order may lead to unexpected behavior.
			cmds = append(cmds, deleteNoteFileCmd(m.listPanel.selectedItem().Path))
			m.listPanel.removeItem()
			cmds = append(cmds, m.renderPreviewCmd(m.listPanel.selectedItem().Path))
			cmd = tea.Batch(cmds...)
		case "ctrl+c":
			m.toggleTypingModal(shut)
		}
	case onFuzzyModal:
		switch msg.String() {
		case "enter":
			m.selectFromFuzzy()
			m.toggleFuzzyModal(shut)
			return m.renderPreviewCmd(m.listPanel.selectedItem().Path)
		case "ctrl+c", "esc":
			m.toggleFuzzyModal(shut)
		case "ctrl+n", "down":
			m.fnsModal.cursorDown()
		case "ctrl+p", "up":
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
		cmd = m.renderPreviewCmd(m.listPanel.selectedItem().Path)
	case tea.KeyMsg:
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
				)))
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
