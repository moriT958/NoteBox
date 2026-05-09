package tui

import (
	"strings"

	"notebox/internal/note"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
	"github.com/sahilm/fuzzy"
)

type filenameSearchModal struct {
	width, height int
	input         textinput.Model
	cursor        int
	offset        int
	filtered      []note.Note
	allItems      []note.Note
}

// filterNotes filters notes by fuzzy matching query against titles.
// Returns all items if query is empty.
func filterNotes(query string, items []note.Note) []note.Note {
	if query == "" {
		return items
	}

	titles := make([]string, len(items))
	for i, n := range items {
		titles[i] = n.Title
	}

	matches := fuzzy.Find(query, titles)
	result := make([]note.Note, len(matches))
	for i, match := range matches {
		result[i] = items[match.Index]
	}

	return result
}

func (m *filenameSearchModal) filter(query string) {
	m.filtered = filterNotes(query, m.allItems)
	m.cursor = 0
	m.offset = 0
}

func (m *filenameSearchModal) cursorUp() {
	m.cursor, m.offset = calcCursorUp(m.cursor, m.offset)
}

func (m *filenameSearchModal) cursorDown() {
	m.cursor, m.offset = calcCursorDown(m.cursor, len(m.filtered), m.offset, m.height)
}

func (m filenameSearchModal) selectedItem() note.Note {
	if len(m.filtered) == 0 {
		return note.Note{}
	}
	return m.filtered[m.cursor]
}

func (m *model) toggleFuzzyModal(ac modalAction) {
	switch ac {
	case open:
		m.fnsModal.input.Reset()
		m.fnsModal.allItems = m.listPanel.items
		m.fnsModal.filtered = m.listPanel.items
		m.fnsModal.cursor = 0
		m.fnsModal.offset = 0
		m.fnsModal.input.Focus()
		m.focus = onFuzzyModal
	case shut:
		m.focus = onListPanel
	}
}

func (m *model) updateFuzzyModalSize(msg tea.WindowSizeMsg) {
	_, v := m.styles.Main.GetFrameSize()
	m.fnsModal.input.Placeholder = "Search notes..."
	m.fnsModal.input.CharLimit = 50
	m.fnsModal.input.SetWidth(m.modalWidth - 4)
	m.fnsModal.width = m.modalWidth
	m.fnsModal.height = (msg.Height - v) / 3
}

func (m *model) selectFromFuzzy() {
	selected := m.fnsModal.selectedItem()
	if selected.Path == "" {
		return
	}

	for i, item := range m.listPanel.items {
		if item.Path == selected.Path {
			m.listPanel.cursor = i
			if m.listPanel.cursor >= m.listPanel.height {
				m.listPanel.offset = m.listPanel.cursor - m.listPanel.height + 1
			} else {
				m.listPanel.offset = 0
			}
			break
		}
	}
}

func (m model) viewFuzzyModal() string {
	filteredList := m.renderFuzzyFilterdList()
	confirm := m.styles.Modal.Confirm.Render(" (" + selectionModalConfirmKey + ") Select ")
	cancel := m.styles.Modal.Cancel.Render(" (" + selectionModalCancelKey + ") Cancel ")
	tip := m.styles.Modal.Centered.
		Width(m.fnsModal.width - 4).
		Render(confirm + "           " + cancel)

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.fnsModal.input.View(),
		"",
		filteredList,
	)

	modalHeight := m.fnsModal.height + 6
	modal := m.styles.Modal.Fuzzy.
		Width(m.fnsModal.width).
		Height(modalHeight).
		Render(content + "\n\n" + tip)
	modal = m.styles.BorderActive.Render(modal)

	background := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewListPanel(),
			m.viewPreviewer(),
		),
		m.viewHelp(),
	)

	overlayX := m.width/2 - m.fnsModal.width/2
	overlayY := m.height/2 - modalHeight/2

	fgLayer := lipgloss.NewLayer(modal).X(overlayX).Y(overlayY).Z(1)
	bgLayer := lipgloss.NewLayer(background).X(0).Y(0).Z(0)

	compositor := lipgloss.NewCompositor(bgLayer, fgLayer)
	canvas := lipgloss.NewCanvas(m.width, m.height).Compose(compositor)

	return m.styles.Main.Render(canvas.Render())
}

func (m *model) renderFuzzyFilterdList() string {
	var listView strings.Builder

	if len(m.fnsModal.filtered) == 0 {
		listView.WriteString("  No matches found")
	} else {
		end := min(m.fnsModal.offset+m.fnsModal.height, len(m.fnsModal.filtered))
		for i := m.fnsModal.offset; i < end; i++ {
			var title string
			if i == m.fnsModal.cursor {
				title = "  " + m.fnsModal.filtered[i].Title
				title = m.styles.Cursor.Render(title)
			} else {
				title = "   " + m.fnsModal.filtered[i].Title
			}
			title = truncate.StringWithTail(title, uint(m.fnsModal.width-4), "...")
			listView.WriteString(title)
			if i != end-1 {
				listView.WriteString("\n")
			}
		}
	}
	return listView.String()
}
