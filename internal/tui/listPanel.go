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

func (m *listPanel) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset--
		}
	}
}

func (m *listPanel) cursorDown() {
	if m.cursor < len(m.items)-1 {
		m.cursor++
		if m.cursor >= m.offset+m.height {
			m.offset++
		}
	}
}

// SelectedItem returns the current selected item in the list.
func (m listPanel) selectedItem() note {
	if m.cursor < 0 || len(m.items) == 0 || len(m.items) <= m.cursor {
		return note{}
	}
	return m.items[m.cursor]
}

// removeItem removes item on current cursor.
func (m *listPanel) removeItem() {
	if m.cursor < 0 || len(m.items) == 0 || len(m.items) <= m.cursor {
		return
	}
	m.items = slices.Delete(m.items, m.cursor, m.cursor+1)

	if m.cursor > len(m.items)-1 {
		m.cursor--
	}
}

func (m *model) updateListPanelSize(msg tea.WindowSizeMsg) {
	m.listPanel.width = msg.Width / 5
	m.listPanel.height = msg.Height * 5 / 6
}

func (m model) viewListPanel() string {
	var view strings.Builder

	if len(m.listPanel.items) == 0 {
		return "no items."
	}

	end := min(m.listPanel.offset+m.height, len(m.listPanel.items))
	for i := m.listPanel.offset; i < end; i++ {
		var title string
		if i == m.listPanel.cursor {
			title = " " + m.listPanel.items[i].title
			title = m.styles.cursorColor.Render(title)
		} else {
			title = "  " + m.listPanel.items[i].title
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
