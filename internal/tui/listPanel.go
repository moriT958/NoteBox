package tui

import (
	"notebox/internal/note"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/muesli/reflow/truncate"
)

type listPanel struct {
	width, height int
	cursor        int
	items         []note.Note
	offset        int

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

const layoutSidePanelRatio = 4

func (m *model) updateListPanelSize(msg tea.WindowSizeMsg) {
	m.listPanel.width = msg.Width / layoutSidePanelRatio
	contentHeight := msg.Height - layoutFramePadding - helpGuideHeight
	m.listPanel.height = max(1, contentHeight)
}

func (m model) viewListPanel() string {
	var view strings.Builder

	if len(m.listPanel.items) == 0 {
		view.WriteString("no items.")
		if m.focus == onListPanel {
			return m.styles.BorderActive.Render(
				m.styles.Sized(m.listPanel.width, m.listPanel.height).Render(view.String()))
		}
		return m.styles.BorderPassive.Render(
			m.styles.Sized(m.listPanel.width, m.listPanel.height).Render(view.String()))
	}

	end := min(m.listPanel.offset+m.listPanel.height, len(m.listPanel.items))
	for i := m.listPanel.offset; i < end; i++ {
		var title string
		if i == m.listPanel.cursor {
			title = "  " + m.listPanel.items[i].Title
			title = m.styles.Cursor.Render(title)
		} else {
			title = "   " + m.listPanel.items[i].Title
		}
		title = truncate.StringWithTail(title, uint(m.listPanel.width), "…   ")
		view.WriteString(title)
		if i != end-1 {
			view.WriteString("\n")
		}
	}

	if m.focus == onListPanel {
		return m.styles.BorderActive.Render(
			m.styles.Sized(m.listPanel.width, m.listPanel.height).Render(view.String()))
	}
	return m.styles.BorderPassive.Render(
		m.styles.Sized(m.listPanel.width, m.listPanel.height).Render(view.String()))
}
