package tui

import (
	"notebox/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type focus int

const (
	onListPanel focus = iota
	onPreviewer
	onTypingModal
	onWarnModal
)

type model struct {
	cfg    *config.Config
	styles *styles

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
}

func NewModel() (*model, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return nil, err
	}

	notes, err := loadNoteFiles(cfg.NotesDir)
	if err != nil {
		return nil, err
	}

	vp := viewport.New(0, 0)
	vp.SetHorizontalStep(4)

	m := &model{
		cfg:         cfg,
		styles:      defaultStyles(),
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
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.SetWindowTitle("NoteBox"),
	)
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
			return m.renderPreviewCmd(m.listPanel.selectedItem().path)
		case "k":
			m.listPanel.cursorUp()
			return m.renderPreviewCmd(m.listPanel.selectedItem().path)
		case "n":
			m.toggleTypingModal(open)
		case "ctrl+l":
			m.focus = onPreviewer
		case "d":
			m.toggleWarnModal(open)
		case "e":
			cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.selectedItem().path)
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
			cmd = openNoteWithEditor(m.cfg.Editor, m.listPanel.selectedItem().path)
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
			cmds = append(cmds, deleteNoteFileCmd(m.listPanel.selectedItem().path))
			m.listPanel.removeItem()
			cmds = append(cmds, m.renderPreviewCmd(m.listPanel.selectedItem().path))
			cmd = tea.Batch(cmds...)
		case "ctrl+c":
			m.toggleTypingModal(shut)
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
		cmd = m.renderPreviewCmd(m.listPanel.selectedItem().path)
	case tea.KeyMsg:
		cmd = m.handleKeyMsg(msg)
	case renderPreviewMsg:
		cmd = m.updatePreviewerContent(msg)
	case newNoteCreatedMsg:
		m.listPanel.items = append(m.listPanel.items, note(msg))
		m.listPanel.cursor = len(m.listPanel.items) - 1
		cmd = m.renderPreviewCmd(m.listPanel.selectedItem().path)
	}

	return m, cmd
}

func (m model) View() string {
	if m.focus == onTypingModal {
		return m.viewTypingModal()
	}
	if m.focus == onWarnModal {
		return m.viewWarnModal()
	}
	view := m.styles.main.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.viewHeader(),
			lipgloss.JoinHorizontal(lipgloss.Top,
				m.viewListPanel(),
				m.viewPreviewer(),
			)))
	return view
}

func (m model) viewHeader() string {
	return m.styles.header.
		Align(lipgloss.Center).
		Width(m.width).
		Render("📓 NoteBox 📓")
}
