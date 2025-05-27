package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"notebox/config"
	stringfunction "notebox/pkg/string_function"
	typingmodal "notebox/typing-modal"

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
	typingModal   typingmodal.Model
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
		typingModal: typingmodal.New(),
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
			m.typingModal, cmd = m.typingModal.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.height, m.width = msg.Height, msg.Width
		m.previewer, cmd = m.previewer.update(msg)
		cmds = append(cmds, cmd)
		m.typingModal, cmd = m.typingModal.Update(msg)
		cmds = append(cmds, cmd)
		m.listPanel, cmd = m.listPanel.update(msg)
		cmds = append(cmds, cmd)
	case typingmodal.TypingModalMsg:
		if bool(msg) {
			m.focus = focusTypingModal
		} else {
			m.focus = focusListPanel
		}
		m.typingModal, cmd = m.typingModal.Update(msg)
		return m, cmd
	case renderPreviewMsg:
		m.previewer, cmd = m.previewer.update(msg)
		return m, cmd
	// case createNewNoteMsg:
	// 	m.listPanel, cmd = m.listPanel.update(msg)
	// 	return m, cmd
	case typingmodal.InputDataMsg:
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
				m.typingModal.View())))
	case focusPreviewer:
		return appStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
			m.appTitleView(),
			lipgloss.JoinHorizontal(lipgloss.Left,
				borderStyle(false).Render(m.listPanel.view()),
				borderStyle(true).Render(m.previewer.view()),
				m.typingModal.View())))
	case focusTypingModal:
		modal := borderStyle(true).Render(m.typingModal.View())
		mw, mh := m.typingModal.GetSize()
		overlayX := m.width/2 - mw/2
		overlayY := m.height/2 - mh/2
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width, m.height = (msg.Width-h)/5, (msg.Height-v)*5/6
		return m, m.renderPreviewCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			return m, typingmodal.ToggleModalCmd(true)
		case "j":
			m.cursorDown()
			return m, m.renderPreviewCmd()
		case "k":
			m.cursorUp()
			return m, m.renderPreviewCmd()
		case "d":
			var cmds []tea.Cmd
			if len(m.items) == 0 {
				break
			}
			cmds = append(cmds, deleteNoteFileCmd(m.selectedItem().title))
			m.removeItem()
			cmds = append(cmds, m.renderPreviewCmd())
			return m, tea.Batch(cmds...)
		case "e":
			note := m.selectedItem()
			return m, openNoteWithEditor(note.title)
		}
	case typingmodal.InputDataMsg:
		title := string(msg)
		return m, m.createNewNoteCmd(title)
	}
	return m, nil
}

func (m *listPanelModel) createNewNoteCmd(title string) tea.Cmd {
	timeStr := time.Now().Format(time.DateOnly)
	filename := filepath.Join(config.BaseDir, title+"-"+timeStr+".md")
	note := note{
		title: title,
		path:  filename,
	}
	newIndex := len(m.items)
	m.items = append(m.items, note)
	m.cursor = newIndex

	return func() tea.Msg {
		fp, err := os.Create(filename)
		if err != nil {
			return errMsg{err}
		}
		defer fp.Close()

		content := fmt.Sprintf("# %s\n\n", title)
		fmt.Fprint(fp, content)

		return nil
	}
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
			str := " " + m.items[i].title
			str = truncate.StringWithTail(str, uint(m.width), "…   ")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(cursorColor)).Render(str))
		} else {
			str := "  " + m.items[i].title
			str = truncate.StringWithTail(str, uint(m.width), "…   ")
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
