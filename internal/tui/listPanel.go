package tui

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"
)

type listPanel struct {
	width, height int
	cursor        int
	items         []note
	offset        int
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
func calcRemoveItem(items []note, cursor int) ([]note, int) {
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

func (m *listPanel) cursorUp() {
	m.cursor, m.offset = calcCursorUp(m.cursor, m.offset)
}

func (m *listPanel) cursorDown() {
	m.cursor, m.offset = calcCursorDown(m.cursor, len(m.items), m.offset, m.height)
}

// Get selected item in the list
func (m listPanel) selectedItem() note {
	if m.cursor < 0 || len(m.items) == 0 || len(m.items) <= m.cursor {
		return note{}
	}
	return m.items[m.cursor]
}

// Add item and adjust cursor/offset to keep it visible
func (m *listPanel) addItem(n note) {
	m.items = append(m.items, n)
	m.cursor, m.offset = calcAddItem(len(m.items), m.offset, m.height)
}

// Remove item on current cursor
func (m *listPanel) removeItem() {
	m.items, m.cursor = calcRemoveItem(m.items, m.cursor)
}

func (m *model) updateListPanelSize(msg tea.WindowSizeMsg) {
	h, v := m.styles.main.GetFrameSize()
	m.listPanel.width = (msg.Width - h) / 5
	m.listPanel.height = (msg.Height - v) * 5 / 6
}

func (m model) viewListPanel() string {
	var view strings.Builder

	if len(m.listPanel.items) == 0 {
		view.WriteString("no items.")
		if m.focus == onListPanel {
			return m.styles.borderActive.Render(
				adjustSize(m.listPanel.width, m.listPanel.height)(view.String()))
		}
		return m.styles.borderPassive.Render(
			adjustSize(m.listPanel.width, m.listPanel.height)(view.String()))
	}

	end := min(m.listPanel.offset+m.listPanel.height, len(m.listPanel.items))
	for i := m.listPanel.offset; i < end; i++ {
		var title string
		if i == m.listPanel.cursor {
			title = "  " + m.listPanel.items[i].title
			title = m.styles.cursorColor.Render(title)
		} else {
			title = "   " + m.listPanel.items[i].title
		}
		title = truncate.StringWithTail(title, uint(m.listPanel.width), "…   ")
		view.WriteString(title)
		if i != end-1 {
			view.WriteString("\n")
		}
	}

	if m.focus == onListPanel {
		return m.styles.borderActive.Render(
			adjustSize(m.listPanel.width, m.listPanel.height)(view.String()))
	}
	return m.styles.borderPassive.Render(
		adjustSize(m.listPanel.width, m.listPanel.height)(view.String()))
}
