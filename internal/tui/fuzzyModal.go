package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"
)

func (m *model) toggleFuzzyModal(ac modalAction) {
	switch ac {
	case open:
		m.fnsModal.Input.Reset()
		m.fnsModal.AllItems = m.listPanel.Items
		m.fnsModal.Filtered = m.listPanel.Items
		m.fnsModal.Cursor = 0
		m.fnsModal.Offset = 0
		m.fnsModal.Input.Focus()
		m.focus = onFuzzyModal
	case shut:
		m.focus = onListPanel
	}
}

func (m *model) updateFuzzyModalSize(msg tea.WindowSizeMsg) {
	_, v := m.styles.Main.GetFrameSize()
	m.fnsModal.Input.Placeholder = "Search notes..."
	m.fnsModal.Input.CharLimit = 50
	m.fnsModal.Input.SetWidth(m.modalWidth - 4)
	m.fnsModal.Width = m.modalWidth
	m.fnsModal.Height = (msg.Height - v) / 3
}

func (m *model) selectFromFuzzy() {
	selected := m.fnsModal.SelectedItem()
	if selected.Path == "" {
		return
	}

	for i, item := range m.listPanel.Items {
		if item.Path == selected.Path {
			m.listPanel.SelectByIndex(i)
			break
		}
	}
}

// Fixed lines in viewFuzzyModal's content around the filtered list: the
// input line, a blank line before the list, a blank line before the tip
// line, and the tip line itself.
const (
	fuzzyModalInputLines     = 1
	fuzzyModalInputToListGap = 1
	fuzzyModalListToTipGap   = 1
	fuzzyModalTipLines       = 1
)

// fuzzyModalHeight is the Height() passed to Modal.Fuzzy in viewFuzzyModal,
// shared with fuzzyModalCursor so the two never drift apart.
func (m model) fuzzyModalHeight() int {
	return m.fnsModal.Height + m.styles.Modal.Fuzzy.GetVerticalPadding() +
		fuzzyModalInputLines + fuzzyModalInputToListGap +
		fuzzyModalListToTipGap + fuzzyModalTipLines
}

func (m model) viewFuzzyModal() string {
	filteredList := m.renderFuzzyFilterdList()
	confirm := m.styles.Modal.Confirm.Render(" (" + selectionModalConfirmKey + ") Select ")
	cancel := m.styles.Modal.Cancel.Render(" (" + selectionModalCancelKey + ") Cancel ")
	tip := m.styles.Modal.Centered.
		Width(m.fnsModal.Width - 4).
		Render(confirm + "           " + cancel)

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.fnsModal.Input.View(),
		"",
		filteredList,
	)

	modalHeight := m.fuzzyModalHeight()
	modal := m.styles.Modal.Fuzzy.
		Width(m.fnsModal.Width).
		Height(modalHeight).
		Render(content + "\n\n" + tip)
	modal = m.styles.BorderActive.Render(modal)

	overlayX := m.width/2 - m.fnsModal.Width/2
	overlayY := m.height/2 - modalHeight/2

	return m.renderOverlay(modal, overlayX, overlayY)
}

func (m model) fuzzyModalCursor() *tea.Cursor {
	cur := m.fnsModal.Input.Cursor()
	if cur == nil {
		return nil
	}
	cur.Position.X = realCursorX(m.fnsModal.Input)
	modalHeight := m.fuzzyModalHeight()
	overlayX := m.width/2 - m.fnsModal.Width/2
	overlayY := m.height/2 - modalHeight/2

	// input.View() is content line 0.
	borderX, borderY := borderSize(m.styles.BorderActive)
	padX, padY := paddingSize(m.styles.Modal.Fuzzy)
	cur.Position.X += overlayX + borderX + padX
	cur.Position.Y += overlayY + borderY + padY
	return cur
}

func (m *model) renderFuzzyFilterdList() string {
	var listView strings.Builder

	if len(m.fnsModal.Filtered) == 0 {
		listView.WriteString("  No matches found")
	} else {
		end := min(m.fnsModal.Offset+m.fnsModal.Height, len(m.fnsModal.Filtered))
		for i := m.fnsModal.Offset; i < end; i++ {
			var title string
			if i == m.fnsModal.Cursor {
				title = "  " + m.fnsModal.Filtered[i].Title
				title = m.styles.Cursor.Render(title)
			} else {
				title = "   " + m.fnsModal.Filtered[i].Title
			}
			title = truncate.StringWithTail(title, uint(m.fnsModal.Width-4), "...")
			listView.WriteString(title)
			if i != end-1 {
				listView.WriteString("\n")
			}
		}
	}
	return listView.String()
}
