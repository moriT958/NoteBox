package main

import (
	"log/slog"
	"os"
	"slices"
	"strings"

	"NoteBox.tmp/config"
	stringfunction "NoteBox.tmp/pkg/string_function"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

/* Main Model */

type model struct {
	height, width int
	focus         int
	listPanel     listPanelModel
	previewer     previewerModel
	typingModal   typingModal
}

const (
	focusListPanel int = iota
	focusPreviewer
	focusTypingModal
)

func newModel() (*model, error) {

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(95),
	)
	if err != nil {
		return nil, err
	}

	notes, err := loadNoteFiles(config.BaseDir)
	if err != nil {
		return nil, err
	}

	m := &model{
		height: 0,
		width:  0,
		focus:  focusListPanel,
		listPanel: listPanelModel{
			items: notes,
		},
		previewer: previewerModel{
			vp:       viewport.New(0, 0),
			renderer: r,
		},
		typingModal: typingModal{
			open:  false,
			input: textinput.New(),
		},
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	// init textinput
	cmds = append(cmds, textinput.Blink)
	// set terminal window title
	cmds = append(cmds, tea.SetWindowTitle("NoteBox"))
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" {
			return m, tea.Quit
		}

		switch m.focus {
		case focusListPanel:
			switch msg.String() {
			case "ctrl+h":
				m.focus = focusListPanel
			case "ctrl+l":
				m.focus = focusPreviewer
			}
			m.listPanel, cmd = m.listPanel.update(msg)
			return m, cmd
		case focusPreviewer:
			switch msg.String() {
			case "ctrl+h":
				m.focus = focusListPanel
			case "ctrl+l":
				m.focus = focusPreviewer
			}
			m.previewer, cmd = m.previewer.update(msg)
			return m, cmd
		case focusTypingModal:
			m.typingModal, cmd = m.typingModal.update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.height, m.width = msg.Height, msg.Width
		m.previewer, cmd = m.previewer.update(msg)
		cmds = append(cmds, cmd)
		m.typingModal, cmd = m.typingModal.update(msg)
		cmds = append(cmds, cmd)
		m.listPanel, cmd = m.listPanel.update(msg)
		cmds = append(cmds, cmd)
	case typingModalMsg:
		if msg.isOpen {
			m.focus = focusTypingModal
		} else {
			m.focus = focusListPanel
		}
		m.typingModal, cmd = m.typingModal.update(msg)
		return m, cmd
	case renderPreviewMsg:
		m.previewer, cmd = m.previewer.update(msg)
		return m, cmd
	case createNewNoteMsg:
		m.listPanel, cmd = m.listPanel.update(msg)
		return m, cmd
	case errMsg:
		slog.Error(msg.err.Error())
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

func (m model) appTitleView() string {
	return appTitleStyle.
		Width(m.width).
		Render("📓 NoteBox 📓")
}

func (m model) View() string {
	switch m.focus {
	case focusListPanel:
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(true).Render(m.listPanel.view()),
				borderStyle(false).Render(m.previewer.view()),
				m.typingModal.view())))
	case focusPreviewer:
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(false).Render(m.listPanel.view()),
				borderStyle(true).Render(m.previewer.view()),
				m.typingModal.view())))
	case focusTypingModal:
		modal := borderStyle(true).Render(m.typingModal.view())
		overlayX := m.width/2 - ModalWidth/2
		overlayY := m.height/2 - ModalHeight/2
		mainView := appStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(false).Render(m.listPanel.view()),
				borderStyle(false).Render(m.previewer.view()))))
		return stringfunction.PlaceOverlay(overlayX, overlayY, modal, mainView)
	default:
		return ""
	}
}

/* List Panel Model */

type listPanelModel struct {
	width, height int
	cursor        int
	items         []note
	selected      string
	offset        int
}

func (m *listPanelModel) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset--
		}
	}
}

func (m *listPanelModel) cursorDown() {
	if m.cursor < len(m.items)-1 {
		m.cursor++
		if m.cursor >= m.offset+m.height {
			m.offset++
		}
	}
}

// SelectedItem returns the current selected item in the list.
func (m listPanelModel) selectedItem() note {
	if m.cursor < 0 || len(m.items) == 0 || len(m.items) <= m.cursor {
		return note{}
	}

	return m.items[m.cursor]
}

func (m *listPanelModel) removeItem() {
	if m.cursor < 0 || len(m.items) == 0 || len(m.items) <= m.cursor {
		return
	}

	m.items = slices.Delete(m.items, m.cursor, m.cursor+1)

	if m.cursor > len(m.items)-1 {
		m.cursor--
	}
}

func (m listPanelModel) update(msg tea.Msg) (listPanelModel, tea.Cmd) {
	var (
		// cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width, m.height = (msg.Width-h)/5, (msg.Height-v)*5/6
		return m, m.renderPreviewCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			return m, toggleModalCmd(true)
		case "j":
			m.cursorDown()
			cmds = append(cmds, m.renderPreviewCmd())
		case "k":
			m.cursorUp()
			cmds = append(cmds, m.renderPreviewCmd())
		case "d":
			if len(m.items) == 0 {
				break
			}
			cmds = append(cmds, deleteNoteFileCmd(m.selectedItem().title))
			m.removeItem()
			cmds = append(cmds, m.renderPreviewCmd())
		case "e":
			note := m.selectedItem()
			return m, openNoteWithEditor(note.title)
		}
	case createNewNoteMsg:
		newIndex := len(m.items)
		newNote := note{
			title: msg.note.title,
			path:  msg.note.path,
		}
		m.items = append(m.items, newNote)
		// TODO:
		// move to this cursor on view
		m.cursor = newIndex

		cmds = append(cmds, m.renderPreviewCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m listPanelModel) renderPreviewCmd() tea.Cmd {
	if len(m.items) > 0 {
		return func() tea.Msg { return renderPreviewMsg{m.selectedItem().path} }
	} else {
		return func() tea.Msg { return renderPreviewMsg{config.DummyNotePath} }
	}
}

func (m listPanelModel) view() string {
	render := lipgloss.NewStyle().Height(m.height).Width(m.width).Render

	var b strings.Builder

	if len(m.items) == 0 {
		return render("No items.")
	}

	end := min(m.offset+m.height, len(m.items))
	for i := m.offset; i < end; i++ {
		if i == m.cursor {
			str := "> " + m.items[i].title
			str = truncate.StringWithTail(str, uint(m.width), "…")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(borderActiveColor)).Render(str))
		} else {
			str := "  " + m.items[i].title
			str = truncate.StringWithTail(str, uint(m.width), "…")
			b.WriteString(str)
		}
		if i != end-1 {
			b.WriteString("\n")
		}
	}

	return render(b.String())
}

/* Previewer Model */
type previewerModel struct {
	vp       viewport.Model
	renderer *glamour.TermRenderer
}

func (m previewerModel) update(msg tea.Msg) (previewerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.vp.Height = msg.Height*5/6 - v
		m.vp.Width = msg.Width*2/3 - h
		m.vp.Height = (msg.Height - v) * 5 / 6
		m.vp.Width = (msg.Width - h) * 2 / 3
	case renderPreviewMsg:
		content, err := os.ReadFile(msg.path)
		if err != nil {
			return m, errCmd(err)
		}
		renderedContent, err := m.renderer.Render(string(content))
		if err != nil {
			return m, errCmd(err)
		}
		m.vp.SetContent(renderedContent)
		m.vp.GotoTop()
	default:
		m.vp, cmd = m.vp.Update(msg)
	}

	return m, cmd
}

func (m previewerModel) view() string {
	view := strings.Builder{}
	view.WriteString(m.vp.View())
	return view.String()
}

/* Typing Modal Model */

type typingModal struct {
	open  bool
	input textinput.Model
}

func (m typingModal) update(msg tea.Msg) (typingModal, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.input.Placeholder = "Enter note name..."
		m.input.Focus()
		m.input.CharLimit = 50
		m.input.Width = msg.Width / 3
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.open = false
			cmds = append(cmds,
				createNewNoteFileCmd(m.input.Value()),
				toggleModalCmd(false))
		case "esc", "ctrl+c":
			m.open = false
			cmds = append(cmds, toggleModalCmd(false))
		}
	case typingModalMsg:
		m.input.Reset()
		m.open = msg.isOpen
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

var (
	ModalConfirm = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#99d1db"))
	ModalCancel  = lipgloss.NewStyle().Foreground(lipgloss.Color("#414559")).Background(lipgloss.Color("#ea999c"))
)

var (
	ModalHeight = 7
	ModalWidth  = 60
)

func (m typingModal) view() string {
	if m.open {
		confirm := ModalConfirm.Render(" (" + "enter" + ") Create ")
		cancel := ModalCancel.Render(" (" + "ctrl+c" + ") Cancel ")

		tip := confirm +
			lipgloss.NewStyle().Render("           ") +
			cancel

		// return ModalBorderStyle(ModalHeight, ModalWidth).Render("\n" + m.input.View() + "\n\n" + tip)
		return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
			Render("\n" + m.input.View() + "\n\n" + tip)
	}
	return ""
}
