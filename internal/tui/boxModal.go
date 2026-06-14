package tui

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"notebox/internal/config"
	"notebox/internal/note"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
)

const (
	titleField = 0
	pathField  = 1
)

type boxModalMode int

const (
	modeNewBox boxModalMode = iota
	modeOpenFolder
)

type boxModal struct {
	width, height int
	cursor        int
	offset        int
	items         []note.Box
	mode          boxModalMode
	titleInput    textinput.Model
	pathInput     textinput.Model
	renameInput   textinput.Model
	activeField   int
	validationErr string
}

func (m *boxModal) cursorUp() {
	m.cursor, m.offset = calcCursorUp(m.cursor, m.offset)
}

func (m *boxModal) cursorDown() {
	m.cursor, m.offset = calcCursorDown(m.cursor, len(m.items), m.offset, m.height)
}

func (m boxModal) selectedItem() note.Box {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return note.Box{}
	}
	return m.items[m.cursor]
}

func (m *model) toggleBoxModal(ac modalAction) {
	switch ac {
	case shut:
		m.focus = onListPanel
	}
}

func (m *model) updateBoxModalSize(msg tea.WindowSizeMsg) {
	_, v := m.styles.Main.GetFrameSize()
	m.boxModal.width = m.modalWidth
	m.boxModal.height = (msg.Height - v) / 3
	m.boxModal.renameInput.SetWidth(m.boxModal.width - 6)
}

func (m *model) switchBox(newBox note.Box) tea.Cmd {
	if err := m.listPanel.registerer.Unregister(m.currentBox.Path); err != nil {
		slog.Error("failed to unregister path", "path", m.currentBox.Path, "error", err)
	}

	ch, err := m.listPanel.registerer.Register(newBox.Path)
	if err != nil {
		slog.Error("failed to register new box path", "path", newBox.Path, "error", err)
		return nil
	}

	m.currentBox = newBox
	m.listPanel.notesUpdates = ch
	m.listPanel.items = []note.Note{}
	m.listPanel.cursor = 0
	m.listPanel.offset = 0
	m.previewer.clearAllTabs()

	return tea.Batch(waitNoteChangeCmd(ch), saveLastBoxCmd(newBox.ID))
}

func (m model) viewBoxModal() string {
	boxList := m.renderBoxList()
	guide := m.help.ShortHelpView([]key.Binding{
		m.keys.boxModal.confirm,
		m.keys.boxModal.cancel,
		m.keys.boxModal.newBox,
		m.keys.boxModal.openFolderAsBox,
		m.keys.boxModal.renameBox,
		m.keys.boxModal.deleteBox,
	})
	tip := m.styles.Help.
		Width(m.boxModal.width - 4).
		Render(guide)

	modalHeight := m.boxModal.height + 4
	modal := m.styles.Modal.Fuzzy.
		Width(m.boxModal.width).
		Height(modalHeight).
		Render(boxList + "\n\n" + tip)
	modal = m.styles.BorderActive.Render(modal)

	overlayX := m.width/2 - m.boxModal.width/2
	overlayY := m.height/2 - modalHeight/2

	return m.renderOverlay(modal, overlayX, overlayY)
}

func (m *model) handleBoxFormConfirm() tea.Cmd {
	title := m.boxModal.titleInput.Value()
	rawPath := m.boxModal.pathInput.Value()

	if title == "" {
		m.boxModal.validationErr = "name is required"
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		m.boxModal.validationErr = "failed to get home directory"
		return nil
	}

	switch m.boxModal.mode {
	case modeNewBox:
		var resolvedBase string
		if rawPath != "" {
			cwd, err := os.Getwd()
			if err != nil {
				m.boxModal.validationErr = "failed to get current directory"
				return nil
			}
			resolvedBase = resolveBoxPath(rawPath, cwd, home)
		}
		finalPath := newBoxFinalPath(title, resolvedBase, filepath.Join(home, config.AppDirName))
		if isDuplicatePath(finalPath, m.boxModal.items) {
			m.boxModal.validationErr = "a box with this path already exists"
			return nil
		}
		m.boxModal.titleInput.Blur()
		m.boxModal.pathInput.Blur()
		m.focus = onListPanel
		return newBoxCmd(m.boxRepo, title, finalPath)

	case modeOpenFolder:
		if rawPath == "" {
			m.boxModal.validationErr = "path is required"
			return nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			m.boxModal.validationErr = "failed to get current directory"
			return nil
		}
		resolved := resolveBoxPath(rawPath, cwd, home)
		info, statErr := os.Stat(resolved)
		if statErr != nil || !info.IsDir() {
			m.boxModal.validationErr = "path must be an existing directory"
			return nil
		}
		if isDuplicatePath(resolved, m.boxModal.items) {
			m.boxModal.validationErr = "a box with this path already exists"
			return nil
		}
		m.boxModal.titleInput.Blur()
		m.boxModal.pathInput.Blur()
		m.focus = onListPanel
		return openFolderAsBoxCmd(m.boxRepo, title, resolved)
	}

	return nil
}

func (m *model) toggleBoxFormModal(ac modalAction, mode boxModalMode) {
	switch ac {
	case open:
		m.boxModal.mode = mode
		m.boxModal.titleInput.Reset()
		m.boxModal.pathInput.Reset()
		m.boxModal.activeField = titleField
		m.boxModal.validationErr = ""
		switch mode {
		case modeNewBox:
			m.boxModal.titleInput.Placeholder = "Box name..."
			m.boxModal.pathInput.Placeholder = "Base path (optional, e.g. ~/projects)..."
		case modeOpenFolder:
			m.boxModal.titleInput.Placeholder = "Display name..."
			m.boxModal.pathInput.Placeholder = "Existing folder path (e.g. ~/projects/work)..."
		}
		m.boxModal.titleInput.Focus()
		m.boxModal.pathInput.Blur()
		m.focus = onBoxCreateModal
	case shut:
		m.boxModal.titleInput.Blur()
		m.boxModal.pathInput.Blur()
		m.focus = onBoxModal
	}
}

func (m *model) updateBoxCreateModalSize(msg tea.WindowSizeMsg) {
	h, _ := m.styles.Main.GetFrameSize()
	inputWidth := (msg.Width - h) / 3
	m.boxModal.titleInput.SetWidth(inputWidth)
	m.boxModal.pathInput.SetWidth(inputWidth)
}

const boxCreateModalHeight = 12

func (m model) viewBoxCreateModal() string {
	var header, actionLabel string
	switch m.boxModal.mode {
	case modeNewBox:
		header = "New Box"
		actionLabel = "Create"
	case modeOpenFolder:
		header = "Open Folder as Box"
		actionLabel = "Open"
	}

	confirm := m.styles.Modal.Confirm.Render(" (" + selectionModalConfirmKey + ") " + actionLabel + " ")
	cancel := m.styles.Modal.Cancel.Render(" (" + selectionModalCancelKey + ") Cancel ")
	tip := confirm + "           " + cancel

	errLine := ""
	if m.boxModal.validationErr != "" {
		errLine = m.styles.Modal.Cancel.Render("  " + m.boxModal.validationErr)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		"Name:  "+m.boxModal.titleInput.View(),
		"",
		"Path:  "+m.boxModal.pathInput.View(),
		errLine,
		"",
		tip,
	)

	modal := m.styles.Modal.Centered.
		Width(m.modalWidth).
		Height(boxCreateModalHeight).
		Render(content)

	modalX := (m.width - m.modalWidth) / 2
	modalY := (m.height - boxCreateModalHeight) / 2
	modal = m.styles.BorderActive.Render(modal)

	return m.renderOverlay(modal, modalX, modalY)
}

func (m model) renderBoxList() string {
	var view strings.Builder

	if len(m.boxModal.items) == 0 {
		view.WriteString("  No boxes found")
		return view.String()
	}

	strikethrough := lipgloss.NewStyle().Strikethrough(true)
	end := min(m.boxModal.offset+m.boxModal.height, len(m.boxModal.items))
	for i := m.boxModal.offset; i < end; i++ {
		b := m.boxModal.items[i]
		var line string
		if i == m.boxModal.cursor && m.focus == onBoxRenaming {
			line = "  " + m.boxModal.renameInput.View()
		} else {
			title := b.Title
			if _, err := os.Stat(b.Path); os.IsNotExist(err) {
				title = strikethrough.Render(title)
			}
			active := ""
			if b.ID == m.currentBox.ID {
				active = " *"
			}
			if i == m.boxModal.cursor {
				line = m.styles.Cursor.Render("  " + title + active)
			} else {
				line = "   " + title + active
			}
		}
		if m.focus != onBoxRenaming || i != m.boxModal.cursor {
			line = truncate.StringWithTail(line, uint(m.boxModal.width-4), "...")
		}
		view.WriteString(line)
		if i != end-1 {
			view.WriteString("\n")
		}
	}

	return view.String()
}
