package tui

import (
	"strings"

	stringfunction "notebox/internal/pkg/string_function"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/sahilm/fuzzy"
)

type fuzzyModal struct {
	width, height int
	input         textinput.Model
	cursor        int
	offset        int
	filtered      []note
	allItems      []note
}

// filterNotes filters notes by fuzzy matching query against titles.
// Returns all items if query is empty.
func filterNotes(query string, items []note) []note {
	if query == "" {
		return items
	}

	titles := make([]string, len(items))
	for i, n := range items {
		titles[i] = n.title
	}

	matches := fuzzy.Find(query, titles)
	result := make([]note, len(matches))
	for i, match := range matches {
		result[i] = items[match.Index]
	}

	return result
}

func (m *fuzzyModal) filter(query string) {
	m.filtered = filterNotes(query, m.allItems)
	m.cursor = 0
	m.offset = 0
}

func (m *fuzzyModal) cursorUp() {
	m.cursor, m.offset = calcCursorUp(m.cursor, m.offset)
}

func (m *fuzzyModal) cursorDown() {
	m.cursor, m.offset = calcCursorDown(m.cursor, len(m.filtered), m.offset, m.height)
}

func (m fuzzyModal) selectedItem() note {
	if len(m.filtered) == 0 {
		return note{}
	}
	return m.filtered[m.cursor]
}

func (m *model) toggleFuzzyModal(ac modalAction) {
	switch ac {
	case open:
		m.fuzzy.input.Reset()
		m.fuzzy.allItems = m.listPanel.items
		m.fuzzy.filtered = m.listPanel.items
		m.fuzzy.cursor = 0
		m.fuzzy.offset = 0
		m.fuzzy.input.Focus()
		m.focus = onFuzzyModal
	case shut:
		m.focus = onListPanel
	}
}

func (m *model) updateFuzzyModalSize(msg tea.WindowSizeMsg) {
	_, v := m.styles.main.GetFrameSize()
	m.fuzzy.input.Placeholder = "Search notes..."
	m.fuzzy.input.CharLimit = 50
	m.fuzzy.input.Width = m.modalWidth - 4
	m.fuzzy.width = m.modalWidth
	m.fuzzy.height = (msg.Height - v) / 3
}

func (m *model) selectFromFuzzy() {
	selected := m.fuzzy.selectedItem()
	if selected.path == "" {
		return
	}

	for i, item := range m.listPanel.items {
		if item.path == selected.path {
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
	var listView strings.Builder

	if len(m.fuzzy.filtered) == 0 {
		listView.WriteString("  No matches found")
	} else {
		end := min(m.fuzzy.offset+m.fuzzy.height, len(m.fuzzy.filtered))
		for i := m.fuzzy.offset; i < end; i++ {
			var title string
			if i == m.fuzzy.cursor {
				title = "  " + m.fuzzy.filtered[i].title
				title = m.styles.cursorColor.Render(title)
			} else {
				title = "   " + m.fuzzy.filtered[i].title
			}
			title = truncate.StringWithTail(title, uint(m.fuzzy.width-4), "...")
			listView.WriteString(title)
			if i != end-1 {
				listView.WriteString("\n")
			}
		}
	}

	confirm := m.styles.modalConfirmColor.Render(" (enter) Select ")
	cancel := m.styles.modalCalcelColor.Render(" (ctrl+c) Cancel ")
	tip := confirm + "           " + cancel

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.fuzzy.input.View(),
		"",
		listView.String(),
	)

	modalHeight := m.fuzzy.height + 6
	modal := lipgloss.NewStyle().
		Width(m.fuzzy.width).
		Height(modalHeight).
		Padding(1, 2).
		Render(content + "\n\n" + tip)

	modal = m.styles.borderActive.Render(modal)
	overlayX := m.width/2 - m.fuzzy.width/2
	overlayY := m.height/2 - modalHeight/2
	background := lipgloss.JoinVertical(lipgloss.Center,
		m.viewHeader(),
		lipgloss.JoinHorizontal(lipgloss.Left,
			m.viewListPanel(),
			m.viewPreviewer()))
	return m.styles.main.Render(
		stringfunction.PlaceOverlay(overlayX, overlayY, modal, background))
}
