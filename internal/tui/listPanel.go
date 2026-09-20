package tui

import (
	"notebox/internal/note"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
)

type listPanel struct {
	width, height int
	// boxRows is the full available content height in terminal rows. It can
	// exceed height*noteItemLines when the content height doesn't divide
	// evenly, and is used to size the border box so it lines up with the
	// previewer panel instead of coming up short.
	boxRows     int
	cursor      int
	items       []note.Note
	offset      int
	renameInput textinput.Model

	// notes dir change watcher
	registerer   note.Registerer
	notesUpdates <-chan []note.Note
}

// Calculates the new cursor and offset when moving up
func calcCursorUp(cursor, offset int) (newCursor, newOffset int) {
	newCursor = cursor
	newOffset = offset
	if cursor > 0 {
		newCursor = cursor - 1
		if newCursor < offset {
			newOffset = offset - 1
		}
	}
	return
}

// Calculates the new cursor and offset when moving down
func calcCursorDown(cursor, itemCount, offset, height int) (newCursor, newOffset int) {
	newCursor = cursor
	newOffset = offset
	if cursor < itemCount-1 {
		newCursor = cursor + 1
		if newCursor >= offset+height {
			newOffset = offset + 1
		}
	}
	return
}

// Calculates the new items and cursor after removing an item
func calcRemoveItem(items []note.Note, cursor int) ([]note.Note, int) {
	if cursor < 0 || len(items) == 0 || cursor >= len(items) {
		return items, cursor
	}
	newItems := slices.Delete(slices.Clone(items), cursor, cursor+1)
	newCursor := cursor
	if newCursor > len(newItems)-1 && newCursor > 0 {
		newCursor--
	}
	return newItems, newCursor
}

// Calculates the new cursor and offset after adding an item
func calcAddItem(itemCount, offset, height int) (newCursor, newOffset int) {
	newCursor = itemCount - 1
	newOffset = offset
	if newCursor >= offset+height {
		newOffset = newCursor - height + 1
	}
	return
}

// Calculates cursor and offset after notes reload while keeping cursor visible.
func preserveSelectionPos(cursor, offset, height, itemCount int) (newCursor, newOffset int) {
	if itemCount <= 0 {
		return 0, 0
	}

	newCursor = cursor
	newOffset = offset

	if newCursor >= itemCount {
		newCursor = itemCount - 1
	}
	if newCursor < 0 {
		newCursor = 0
	}

	if height <= 0 {
		newOffset = 0
		return
	}
	if newCursor < newOffset {
		newOffset = newCursor
	}
	if newCursor >= newOffset+height {
		newOffset = newCursor - height + 1
	}
	if newOffset < 0 {
		newOffset = 0
	}
	return
}

func (m *listPanel) cursorUp() {
	m.cursor, m.offset = calcCursorUp(m.cursor, m.offset)
}

func (m *listPanel) cursorDown() {
	m.cursor, m.offset = calcCursorDown(m.cursor, len(m.items), m.offset, m.height)
}

// SelectByIndex jumps the cursor to item index i and repositions the
// viewport window to it. This is the "select this exact item" seam other
// modals (fuzzy search, box switch) go through instead of recomputing
// cursor/offset positioning themselves.
func (m *listPanel) SelectByIndex(i int) {
	if i < 0 || i >= len(m.items) {
		return
	}
	m.cursor = i
	if m.cursor >= m.height {
		m.offset = m.cursor - m.height + 1
	} else {
		m.offset = 0
	}
}

// Get selected item in the list
func (m listPanel) selectedItem() note.Note {
	if m.cursor < 0 || len(m.items) == 0 || len(m.items) <= m.cursor {
		return note.Note{}
	}
	return m.items[m.cursor]
}

// Add item and adjust cursor/offset to keep it visible
func (m *listPanel) addItem(n note.Note) {
	m.items = append(m.items, n)
	m.cursor, m.offset = calcAddItem(len(m.items), m.offset, m.height)
}

// Remove item on current cursor
func (m *listPanel) removeItem() {
	m.items, m.cursor = calcRemoveItem(m.items, m.cursor)
}

// reloadAllNotes reloads all notes
func (m *model) reloadAllNotes(notes []note.Note) {
	selectedPath := m.listPanel.selectedItem().Path
	m.listPanel.items = notes

	for i, n := range notes {
		if n.Path == selectedPath {
			m.listPanel.cursor = i
			break
		}
	}

	m.listPanel.cursor, m.listPanel.offset = preserveSelectionPos(
		m.listPanel.cursor,
		m.listPanel.offset,
		m.listPanel.height,
		len(notes),
	)
}

const layoutListPanelRatio = 4

// Each note item occupies three rows: the title, the path below it, and a
// blank spacer row separating it from the next item.
const noteItemLines = 3

func (m *model) updateListPanelSize(msg tea.WindowSizeMsg) {
	m.listPanel.width = msg.Width / layoutListPanelRatio

	_, borderV := m.styles.BorderPassive.GetFrameSize()
	contentHeight := msg.Height - borderV - helpGuideHeight - headerHeight

	m.listPanel.boxRows = max(1, contentHeight)
	m.listPanel.height = max(1, contentHeight/noteItemLines)
	m.listPanel.renameInput.SetWidth(m.listPanel.width - 4)

	m.listPanel.cursor, m.listPanel.offset = preserveSelectionPos(
		m.listPanel.cursor,
		m.listPanel.offset,
		m.listPanel.height,
		len(m.listPanel.items),
	)
}

func (m model) viewListPanel() string {
	var view strings.Builder

	if len(m.listPanel.items) == 0 {
		view.WriteString("no items.")
		return m.renderListPanelWithBorder(view.String())
	}

	end := min(m.listPanel.offset+m.listPanel.height, len(m.listPanel.items))
	for i, n := range m.listPanel.items[m.listPanel.offset:end] {
		item := m.renderNoteItemLine(n)
		view.WriteString(item)
		if i != end-m.listPanel.offset-1 {
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
	if n != m.listPanel.selectedItem() {
		titleLine := plainGutter + n.Title
		titleLine = truncate.StringWithTail(titleLine, uint(m.listPanel.width), "…   ")
		return titleLine + "\n" + m.renderNotePathLine(n, plainGutter, false) + "\n"
	}

	gutter := m.styles.Cursor.Render(gutterBar) + " "

	if m.focus == onRenaming {
		titleLine := gutter + m.listPanel.renameInput.View()
		return titleLine + "\n" + m.renderNotePathLine(n, gutter, true) + "\n"
	}

	titleLine := gutter + m.styles.Cursor.Render(n.Title)
	titleLine = truncate.StringWithTail(titleLine, uint(m.listPanel.width), "…   ")
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
	return truncate.StringWithTail(line, uint(m.listPanel.width), "…")
}

func (m model) listRenameCursor() *tea.Cursor {
	cur := m.listPanel.renameInput.Cursor()
	if cur == nil {
		return nil
	}
	cur.Position.X = realCursorX(m.listPanel.renameInput)
	rowIdx := (m.listPanel.cursor - m.listPanel.offset) * noteItemLines

	// list panel sits at x=0 in the header/list/help stack.
	borderX, borderY := borderSize(m.styles.BorderActive)
	cur.Position.X += borderX + lipgloss.Width(renameRowPrefix)
	cur.Position.Y += headerHeight + borderY + rowIdx
	return cur
}

func (m model) renderListPanelWithBorder(content string) string {
	rows := m.listPanel.boxRows
	if m.focus == onListPanel {
		return m.styles.BorderActive.Render(
			m.styles.Sized(m.listPanel.width, rows).Render(content),
		)
	}
	return m.styles.BorderPassive.Render(
		m.styles.Sized(m.listPanel.width, rows).Render(content),
	)
}
