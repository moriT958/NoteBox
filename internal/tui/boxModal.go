package tui

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"notebox/internal/config"
	"notebox/internal/notedeprecated"
	"notebox/internal/tui/boxmodal"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
)

const (
	boxCreateNameLabel = "Name:  "
	boxCreatePathLabel = "Path:  "
	boxRenameRowPrefix = "  "
)

func (m *model) toggleBoxModal(ac modalAction) {
	switch ac {
	case shut:
		m.focus = onListPanel
	}
}

func (m *model) updateBoxModalSize(msg tea.WindowSizeMsg) {
	_, v := m.styles.Main.GetFrameSize()
	m.boxModal.Width = m.modalWidth
	m.boxModal.Height = (msg.Height - v) / 3
	m.boxModal.RenameInput.SetWidth(m.boxModal.Width - 6)
}

func (m *model) switchBox(newBox notedeprecated.Box) tea.Cmd {
	if err := m.listPanel.Registerer.Unregister(m.currentBox.Path); err != nil {
		slog.Error("failed to unregister path", "path", m.currentBox.Path, "error", err)
	}

	ch, err := m.listPanel.Registerer.Register(newBox.Path)
	if err != nil {
		slog.Error("failed to register new box path", "path", newBox.Path, "error", err)
		return nil
	}

	m.currentBox = newBox
	m.listPanel.Reset(ch)
	m.previewer.ClearAllTabs()

	return tea.Batch(waitNoteChangeCmd(ch), saveLastBoxCmd(newBox.ID))
}

// boxModalListToTipGap/boxModalTipLines are the fixed lines below boxList in
// viewBoxModal's content: one blank separator line, then the tip line.
const (
	boxModalListToTipGap = 1
	boxModalTipLines     = 1
)

// boxModalHeight is the Height() passed to Modal.Fuzzy in viewBoxModal,
// shared with boxRenameCursor so the two never drift apart.
func (m model) boxModalHeight() int {
	return m.boxModal.Height + m.styles.Modal.Fuzzy.GetVerticalPadding() +
		boxModalListToTipGap + boxModalTipLines
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
		Width(m.boxModal.Width - 4).
		Render(guide)

	modalHeight := m.boxModalHeight()
	modal := m.styles.Modal.Fuzzy.
		Width(m.boxModal.Width).
		Height(modalHeight).
		Render(boxList + "\n\n" + tip)
	modal = m.styles.BorderActive.Render(modal)

	overlayX := m.width/2 - m.boxModal.Width/2
	overlayY := m.height/2 - modalHeight/2

	return m.renderOverlay(modal, overlayX, overlayY)
}

func (m *model) handleBoxFormConfirm() tea.Cmd {
	title := m.boxModal.TitleInput.Value()
	rawPath := m.boxModal.PathInput.Value()

	if title == "" {
		m.boxModal.ValidationErr = "name is required"
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		m.boxModal.ValidationErr = "failed to get home directory"
		return nil
	}

	switch m.boxModal.Mode {
	case boxmodal.ModeNewBox:
		var resolvedBase string
		if rawPath != "" {
			cwd, err := os.Getwd()
			if err != nil {
				m.boxModal.ValidationErr = "failed to get current directory"
				return nil
			}
			resolvedBase = resolveBoxPath(rawPath, cwd, home)
		}
		finalPath := newBoxFinalPath(title, resolvedBase, filepath.Join(home, config.AppDirName))
		if isDuplicatePath(finalPath, m.boxModal.Items) {
			m.boxModal.ValidationErr = "a box with this path already exists"
			return nil
		}
		m.boxModal.TitleInput.Blur()
		m.boxModal.PathInput.Blur()
		m.focus = onListPanel
		return newBoxCmd(m.boxRepo, title, finalPath)

	case boxmodal.ModeOpenFolder:
		if rawPath == "" {
			m.boxModal.ValidationErr = "path is required"
			return nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			m.boxModal.ValidationErr = "failed to get current directory"
			return nil
		}
		resolved := resolveBoxPath(rawPath, cwd, home)
		info, statErr := os.Stat(resolved)
		if statErr != nil || !info.IsDir() {
			m.boxModal.ValidationErr = "path must be an existing directory"
			return nil
		}
		if isDuplicatePath(resolved, m.boxModal.Items) {
			m.boxModal.ValidationErr = "a box with this path already exists"
			return nil
		}
		m.boxModal.TitleInput.Blur()
		m.boxModal.PathInput.Blur()
		m.focus = onListPanel
		return openFolderAsBoxCmd(m.boxRepo, title, resolved)
	}

	return nil
}

func (m *model) toggleBoxFormModal(ac modalAction, mode boxmodal.Mode) {
	switch ac {
	case open:
		m.boxModal.Mode = mode
		m.boxModal.TitleInput.Reset()
		m.boxModal.PathInput.Reset()
		m.boxModal.ActiveField = boxmodal.TitleField
		m.boxModal.ValidationErr = ""
		switch mode {
		case boxmodal.ModeNewBox:
			m.boxModal.TitleInput.Placeholder = "Box name..."
			m.boxModal.PathInput.Placeholder = "Base path (optional, e.g. ~/projects)..."
		case boxmodal.ModeOpenFolder:
			m.boxModal.TitleInput.Placeholder = "Display name..."
			m.boxModal.PathInput.Placeholder = "Existing folder path (e.g. ~/projects/work)..."
		}
		m.boxModal.TitleInput.Focus()
		m.boxModal.PathInput.Blur()
		m.focus = onBoxCreateModal
	case shut:
		m.boxModal.TitleInput.Blur()
		m.boxModal.PathInput.Blur()
		m.focus = onBoxModal
	}
}

func (m *model) updateBoxCreateModalSize(msg tea.WindowSizeMsg) {
	h, _ := m.styles.Main.GetFrameSize()
	inputWidth := min((msg.Width-h)/3, m.modalWidth/2-7)
	m.boxModal.TitleInput.SetWidth(inputWidth)
	m.boxModal.PathInput.SetWidth(inputWidth)
}

const boxCreateModalHeight = 12

const (
	boxCreateModalTitleLine = 2
	boxCreateModalPathLine  = 4
)

func (m model) boxCreateModalLines() []string {
	var header, actionLabel string
	switch m.boxModal.Mode {
	case boxmodal.ModeNewBox:
		header = "New Box"
		actionLabel = "Create"
	case boxmodal.ModeOpenFolder:
		header = "Open Folder as Box"
		actionLabel = "Open"
	}

	confirm := m.styles.Modal.Confirm.Render(" (" + selectionModalConfirmKey + ") " + actionLabel + " ")
	cancel := m.styles.Modal.Cancel.Render(" (" + selectionModalCancelKey + ") Cancel ")
	tip := confirm + "           " + cancel

	errLine := ""
	if m.boxModal.ValidationErr != "" {
		errLine = m.styles.Modal.Cancel.Render("  " + m.boxModal.ValidationErr)
	}

	lines := make([]string, 8)
	lines[0] = header
	lines[1] = ""
	lines[boxCreateModalTitleLine] = boxCreateNameLabel + m.boxModal.TitleInput.View()
	lines[3] = ""
	lines[boxCreateModalPathLine] = boxCreatePathLabel + m.boxModal.PathInput.View()
	lines[5] = errLine
	lines[6] = ""
	lines[7] = tip
	return lines
}

func (m model) viewBoxCreateModal() string {
	content := lipgloss.JoinVertical(lipgloss.Left, m.boxCreateModalLines()...)

	modal := m.styles.Modal.Centered.
		Width(m.modalWidth).
		Height(boxCreateModalHeight).
		Render(content)

	modalX := (m.width - m.modalWidth) / 2
	modalY := (m.height - boxCreateModalHeight) / 2
	modal = m.styles.BorderActive.Render(modal)

	return m.renderOverlay(modal, modalX, modalY)
}

func (m model) boxCreateModalCursor() *tea.Cursor {
	var cur *tea.Cursor
	var lineIdx int
	var prefixWidth int
	switch m.boxModal.ActiveField {
	case boxmodal.TitleField:
		cur = m.boxModal.TitleInput.Cursor()
		if cur != nil {
			cur.Position.X = realCursorX(m.boxModal.TitleInput)
		}
		lineIdx = boxCreateModalTitleLine
		prefixWidth = lipgloss.Width(boxCreateNameLabel)
	case boxmodal.PathField:
		cur = m.boxModal.PathInput.Cursor()
		if cur != nil {
			cur.Position.X = realCursorX(m.boxModal.PathInput)
		}
		lineIdx = boxCreateModalPathLine
		prefixWidth = lipgloss.Width(boxCreatePathLabel)
	}
	if cur == nil {
		return nil
	}

	// viewBoxCreateModal joins these lines with lipgloss.JoinVertical(Left, ...),
	// which right-pads every line to the widest one before centering, so all
	// lines share the same left offset based on the widest line, not their own.
	lines := m.boxCreateModalLines()
	widest := lines[0]
	for _, l := range lines {
		if lipgloss.Width(l) > lipgloss.Width(widest) {
			widest = l
		}
	}
	x, y := centeredLinePos(m.modalWidth, boxCreateModalHeight, len(lines), lineIdx, widest)

	modalX := (m.width - m.modalWidth) / 2
	modalY := (m.height - boxCreateModalHeight) / 2
	borderX, borderY := borderSize(m.styles.BorderActive)
	cur.Position.X += modalX + borderX + x + prefixWidth
	cur.Position.Y += modalY + borderY + y
	return cur
}

func (m model) boxRenameCursor() *tea.Cursor {
	cur := m.boxModal.RenameInput.Cursor()
	if cur == nil {
		return nil
	}
	cur.Position.X = realCursorX(m.boxModal.RenameInput)
	modalHeight := m.boxModalHeight()
	overlayX := m.width/2 - m.boxModal.Width/2
	overlayY := m.height/2 - modalHeight/2
	rowIdx := m.boxModal.Cursor - m.boxModal.Offset

	borderX, borderY := borderSize(m.styles.BorderActive)
	padX, padY := paddingSize(m.styles.Modal.Fuzzy)
	cur.Position.X += overlayX + borderX + padX + lipgloss.Width(boxRenameRowPrefix)
	cur.Position.Y += overlayY + borderY + padY + rowIdx
	return cur
}

func (m model) renderBoxList() string {
	var view strings.Builder

	if len(m.boxModal.Items) == 0 {
		view.WriteString("  No boxes found")
		return view.String()
	}

	strikethrough := lipgloss.NewStyle().Strikethrough(true)
	end := min(m.boxModal.Offset+m.boxModal.Height, len(m.boxModal.Items))
	for i := m.boxModal.Offset; i < end; i++ {
		b := m.boxModal.Items[i]
		var line string
		if i == m.boxModal.Cursor && m.focus == onBoxRenaming {
			line = boxRenameRowPrefix + m.boxModal.RenameInput.View()
		} else {
			title := b.Title
			if _, err := os.Stat(b.Path); os.IsNotExist(err) {
				title = strikethrough.Render(title)
			}
			active := ""
			if b.ID == m.currentBox.ID {
				active = " *"
			}
			if i == m.boxModal.Cursor {
				line = m.styles.Cursor.Render("  " + title + active)
			} else {
				line = "   " + title + active
			}
		}
		if m.focus != onBoxRenaming || i != m.boxModal.Cursor {
			line = truncate.StringWithTail(line, uint(m.boxModal.Width-4), "...")
		}
		view.WriteString(line)
		if i != end-1 {
			view.WriteString("\n")
		}
	}

	return view.String()
}
