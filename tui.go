package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"NoteBox.tmp/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
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
		glamour.WithWordWrap(40),
	)
	if err != nil {
		return nil, err
	}

	notes, err := loadNoteFiles(config.BaseDir)
	if err != nil {
		return nil, err
	}
	items := make([]list.Item, 0)
	for i := range notes {
		items = append(items, notes[i])
	}
	l := list.New(items, itemDelegate{}, 0, 20)
	l.Title = "Your Notes"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := &model{
		height: 0,
		width:  0,
		focus:  focusListPanel,
		listPanel: listPanelModel{
			list: l,
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
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Top,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(true).Render(m.listPanel.view()),
				borderStyle(false).Render(m.previewer.view()),
				m.typingModal.view())))
	case focusPreviewer:
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Top,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(false).Render(m.listPanel.view()),
				borderStyle(true).Render(m.previewer.view()),
				m.typingModal.view())))
	default:
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Top,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(false).Render(m.listPanel.view()),
				borderStyle(false).Render(m.previewer.view()),
				borderStyle(true).Render(m.typingModal.view()))))
	}
}

/* List Panel Model */

func (n note) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	n, ok := listItem.(note)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, n.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type listPanelModel struct {
	list list.Model
}

func (m listPanelModel) update(msg tea.Msg) (listPanelModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width*2/3-h, msg.Height*5/6-v)
		// m.list.SetWidth(msg.Width / 3)
		// m.list.SetHeight(msg.width)
		return m, m.renderPreviewCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			return m, toggleModalCmd(true)
		case "d":
			if len(m.list.Items()) == 0 {
				break
			}
			cmds = append(cmds, deleteNoteFile(m.list.SelectedItem().(note).title))
			m.list.RemoveItem(m.list.Cursor())
			m.list.CursorUp()
			cmds = append(cmds, m.renderPreviewCmd())
		case "e":
			note := m.list.SelectedItem().(note)
			return m, openNoteWithEditor(note.title)
		default:
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			cmds = append(cmds, m.renderPreviewCmd())
		}
	case createNewNoteMsg:
		newIndex := len(m.list.Items())
		newNote := note{
			title: msg.note.title,
			path:  msg.note.path,
		}
		cmd = m.list.InsertItem(newIndex, newNote)
		cmds = append(cmds, cmd)

		m.list.Select(newIndex)
		cmds = append(cmds, m.renderPreviewCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m listPanelModel) renderPreviewCmd() tea.Cmd {
	if len(m.list.Items()) > 0 {
		return func() tea.Msg { return renderPreviewMsg{m.list.SelectedItem().(note).path} }
	} else {
		return func() tea.Msg { return renderPreviewMsg{config.DummyNotePath} }
	}
}

func (m listPanelModel) view() string {
	view := strings.Builder{}
	view.WriteString(m.list.View())
	return view.String()
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

func (m typingModal) view() string {
	var view strings.Builder
	if m.open {
		view.WriteString(m.input.View())
	}
	return view.String()
}
