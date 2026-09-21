package tui

import (
	"notebox/internal/note"
	"notebox/internal/tui/listpanel"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
)

const layoutListPanelRatio = 4

func (m *model) updateListPanelSize(msg tea.WindowSizeMsg) {
	_, borderV := m.styles.BorderPassive.GetFrameSize()
	contentHeight := msg.Height - borderV - helpGuideHeight - headerHeight

	width := msg.Width / layoutListPanelRatio
	boxRows := max(1, contentHeight)
	height := max(1, contentHeight/listpanel.NoteItemLines)
	m.listPanel.SetSize(width, boxRows, height)
}

func (m model) viewListPanel() string {
	var view strings.Builder

	if len(m.listPanel.Items) == 0 {
		view.WriteString("no items.")
		return m.renderListPanelWithBorder(view.String())
	}

	end := min(m.listPanel.Offset+m.listPanel.Height, len(m.listPanel.Items))
	for i, n := range m.listPanel.Items[m.listPanel.Offset:end] {
		item := m.renderNoteItemLine(n)
		view.WriteString(item)
		if i != end-m.listPanel.Offset-1 {
			view.WriteString("\n")
		}
	}

	return m.renderListPanelWithBorder(view.String())
}

// gutterBar is the glow-style marker shown on the selected row; renameRowPrefix
// is its plain-width equivalent used only to line up the rename text cursor.
// The bar is rendered in a single fixed color on both the title and path lines
// so it doesn't change color between them; only the text next to it does.
const (
	gutterBar       = "│"
	renameRowPrefix = gutterBar + " "
	plainGutter     = "  "
)

func (m model) renderNoteItemLine(n note.Note) string {
	if n != m.listPanel.SelectedItem() {
		titleLine := plainGutter + n.Title
		titleLine = truncate.StringWithTail(titleLine, uint(m.listPanel.Width), "…   ")
		return titleLine + "\n" + m.renderNotePathLine(n, plainGutter, false) + "\n"
	}

	gutter := m.styles.Cursor.Render(gutterBar) + " "

	if m.focus == onRenaming {
		titleLine := gutter + m.listPanel.RenameInput.View()
		return titleLine + "\n" + m.renderNotePathLine(n, gutter, true) + "\n"
	}

	titleLine := gutter + m.styles.Cursor.Render(n.Title)
	titleLine = truncate.StringWithTail(titleLine, uint(m.listPanel.Width), "…   ")
	return titleLine + "\n" + m.renderNotePathLine(n, gutter, true) + "\n"
}

// renderNotePathLine renders the note's path, relative to the current box, below
// the title. Selected rows pick up the cursor highlight color; others stay dim.
func (m model) renderNotePathLine(n note.Note, gutter string, selected bool) string {
	path := n.Path
	if rel, err := filepath.Rel(m.currentBox.Path, n.Path); err == nil {
		path = rel
	}

	style := m.styles.NotePath
	if selected {
		style = m.styles.CursorPath
	}

	line := gutter + style.Render(path)
	return truncate.StringWithTail(line, uint(m.listPanel.Width), "…")
}

func (m model) listRenameCursor() *tea.Cursor {
	cur := m.listPanel.RenameInput.Cursor()
	if cur == nil {
		return nil
	}
	cur.Position.X = realCursorX(m.listPanel.RenameInput)
	rowIdx := (m.listPanel.Cursor - m.listPanel.Offset) * listpanel.NoteItemLines

	// list panel sits at x=0 in the header/list/help stack.
	borderX, borderY := borderSize(m.styles.BorderActive)
	cur.Position.X += borderX + lipgloss.Width(renameRowPrefix)
	cur.Position.Y += headerHeight + borderY + rowIdx
	return cur
}

func (m model) renderListPanelWithBorder(content string) string {
	rows := m.listPanel.BoxRows
	if m.focus == onListPanel {
		return m.styles.BorderActive.Render(
			m.styles.Sized(m.listPanel.Width, rows).Render(content),
		)
	}
	return m.styles.BorderPassive.Render(
		m.styles.Sized(m.listPanel.Width, rows).Render(content),
	)
}
