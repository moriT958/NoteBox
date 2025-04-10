package tui

import (
	"fmt"
	"log"
	"notebox/config"
	"notebox/models"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	docStyle  = lipgloss.NewStyle().Margin(0, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
)

type (
	errMsg            struct{ error }
	editorFinishedMsg struct{ error }
)

type focusedArea int

const (
	focusListArea focusedArea = iota
	focusPreviewArea
)

type mode int

const (
	createMode mode = iota
	navigateMode
)

type model struct {
	focus     focusedArea
	mode      mode
	warnModal warnModal
	listPanel listPanel
	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	input     textinput.Model
}

func initModel() (*model, error) {
	// Get notes from repository
	notes, err := models.GetRepository().FindAll()
	if err != nil {
		return nil, err
	}

	const (
		height int = 45
		width  int = 120
	)

	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	const glamourGutter = 2
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}

	ti := textinput.New()
	ti.Placeholder = "New note name..."
	ti.Prompt = "$ "
	ti.CharLimit = 50
	ti.Width = 30

	m := &model{
		listPanel: listPanel{
			height: 45,
			width:  10,
			notes:  notes,
			cursor: 0,
		},
		focus: focusListArea,
		mode:  navigateMode,
		warnModal: warnModal{
			open:    false,
			message: "",
		},
		viewport: vp,
		renderer: renderer,
		input:    ti,
	}

	m.viewport.SetContent("(No note selected)")
	m.viewport.GotoTop()

	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("NoteBox"),
	)
}

func (m *model) openFileWithEditor(file string) tea.Cmd {
	editor := config.Editor()
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	parts := strings.Fields(editor)
	cmd := parts[0]
	args := append(parts[1:], file)

	c := exec.Command(cmd, args...)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case editorFinishedMsg:
		note := m.listPanel.notes[m.listPanel.cursor]
		content, err := os.ReadFile(note.GetFilePath())
		if err != nil {
			m.Update(errMsg{err})
		}
		str, err := m.renderer.Render(string(content))
		if err != nil {
			m.Update(errMsg{err})
		}
		m.viewport.SetContent(str)
		m.viewport.GotoTop()

	case tea.KeyMsg:
		if m.input.Focused() {
			switch msg.String() {
			case "enter":
				if m.mode == createMode {
					title := m.input.Value()
					note := &models.Note{
						Title:    title,
						CreateAt: time.Now(),
					}
					id, err := models.GetRepository().Save(*note)
					if err != nil {
						cmd = func() tea.Msg { return errMsg{err} }
					}

					note, err = models.GetRepository().FindByID(id)
					if err != nil {
						return m.Update(errMsg{err})
					}

					topHeader := "# " + title + "\n\n"
					fp, err := os.Create(note.GetFilePath())
					if err != nil {
						return m.Update(errMsg{err})
					}
					defer fp.Close()
					fmt.Fprint(fp, topHeader)

					m.listPanel.notes = append(m.listPanel.notes, note)
					m.listPanel.cursor = len(m.listPanel.notes) - 1

					content, err := os.ReadFile(note.GetFilePath())
					if err != nil {
						m.Update(errMsg{err})
					}

					str, err := m.renderer.Render(string(content))
					if err != nil {
						m.Update(errMsg{err})
					}

					m.viewport.SetContent(str)
					m.viewport.GotoTop()

				}
				m.input.SetValue("")
				m.input.Blur()
			case "esc", "ctrl+c":
				m.input.SetValue("")
				m.mode = navigateMode
				m.input.Blur()
			}
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)

		} else if m.warnModal.open {
			switch msg.String() {
			case "y":
				if err := os.Remove(m.listPanel.notes[m.listPanel.cursor].GetFilePath()); err != nil {
					m.Update(errMsg{err})
				}

				id := m.listPanel.notes[m.listPanel.cursor].ID
				models.GetRepository().DeleteByID(id)
				m.listPanel.notes = slices.Delete(m.listPanel.notes, m.listPanel.cursor, m.listPanel.cursor+1)
				if m.listPanel.cursor > 0 && m.listPanel.cursor >= len(m.listPanel.notes) {
					m.listPanel.cursor--
				}
				m.warnModal.open = false
				m.warnModal.message = ""

				note := m.listPanel.notes[m.listPanel.cursor]
				content, err := os.ReadFile(note.GetFilePath())
				if err != nil {
					m.Update(errMsg{err})
				}

				str, err := m.renderer.Render(string(content))
				if err != nil {
					m.Update(errMsg{err})
				}

				m.viewport.SetContent(str)
				m.viewport.GotoTop()

			case "n", "esc":
				m.warnModal.open = false
				m.warnModal.message = ""
			}
			return m, nil
		} else if m.focus == focusListArea {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit

			case "e":
				id := m.listPanel.notes[m.listPanel.cursor].ID
				note, err := models.GetRepository().FindByID(id)
				if err != nil {
					log.Println("note does't exit:", err)
				}
				cmd = m.openFileWithEditor(note.GetFilePath())

			case "n":
				m.mode = createMode
				m.input.Focus()
				cmd = textinput.Blink

			case "d":
				m.deleteNoteWarn()

			case "down", "j":
				m.listPanel.cursor++
				if m.listPanel.cursor >= len(m.listPanel.notes) {
					m.listPanel.cursor = 0
				}

				id := m.listPanel.notes[m.listPanel.cursor].ID
				note, err := models.GetRepository().FindByID(id)
				if err != nil {
					log.Fatalln(err)
				}
				content, err := os.ReadFile(note.GetFilePath())
				if err != nil {
					log.Fatalln(err)
				}

				str, err := m.renderer.Render(string(content))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				m.viewport.SetContent(str)
				m.viewport.GotoTop()

			case "up", "k":
				m.listPanel.cursor--
				if m.listPanel.cursor < 0 {
					m.listPanel.cursor = len(m.listPanel.notes) - 1
				}

				id := m.listPanel.notes[m.listPanel.cursor].ID
				note, err := models.GetRepository().FindByID(id)
				if err != nil {
					log.Println(err)
				}
				content, err := os.ReadFile(note.GetFilePath())
				if err != nil {
					log.Println(err)
				}

				str, err := m.renderer.Render(string(content))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				m.viewport.SetContent(str)
				m.viewport.GotoTop()

			case "right", "ctrl+l":
				m.focus = focusPreviewArea
			}
			cmds = append(cmds, cmd)
		} else if m.focus == focusPreviewArea {
			switch msg.String() {
			case "left", "ctrl+h":
				m.focus = focusListArea

			case "ctrl+c", "q", "esc":
				return m, tea.Quit

			default:
				m.viewport, cmd = m.viewport.Update(msg)
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) listPanelRender() string {
	s := strings.Builder{}
	s.WriteString("📓 Your Notes 📓\n\n")
	for i := 0; i < len(m.listPanel.notes); i++ {
		if m.listPanel.cursor == i {
			s.WriteString(lipgloss.NewStyle().
				Reverse(true).
				Render(m.listPanel.notes[i].Title))
		} else {
			s.WriteString(m.listPanel.notes[i].Title)
			s.WriteString(" ")
		}
		s.WriteString("\n")

	}

	return lipgloss.NewStyle().
		Margin(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).Render(s.String())
}

func (m model) View() string {
	final := strings.Builder{}

	listPanel := m.listPanelRender()
	final.WriteString(listPanel)

	if m.input.Focused() {
		prompt := "Enter note name:\n\n" + m.input.View()
		final.WriteString("\n" + docStyle.Render(prompt) + "\n")
		return final.String()
	}

	if m.warnModal.open {
		final.WriteString("\n" + docStyle.Render(m.warnModal.message) + "\n")
		return final.String()
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, final.String(), m.viewport.View()+m.helpView())
}

func (m model) helpView() string {
	if m.focus == focusListArea {
		return helpStyle.Render("\n ↑/↓: Navigate • e: Edit • ctrl+l: focus-preview • q: Quit")
	} else if m.focus == focusPreviewArea {
		return helpStyle.Render("\n  ↑/↓: Navigate • ctrl+h: focus-list • q: Quit\n")
	}
	return ""
}

func (m *model) deleteNoteWarn() {
	m.warnModal = warnModal{
		open:    true,
		message: "Are you sure you want to completely delete?(y/n)",
	}
}
