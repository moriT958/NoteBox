// Package fuzzymodal owns the fuzzy filename search modal's state: its
// filtered result list and fuzzy-matching/cursor operations.
package fuzzymodal

import (
	"notebox/internal/notedeprecated"

	"charm.land/bubbles/v2/textinput"
	"github.com/sahilm/fuzzy"
)

type Modal struct {
	Width, Height int
	Input         textinput.Model
	Cursor        int
	Offset        int
	Filtered      []notedeprecated.Note
	AllItems      []notedeprecated.Note
}

// filterNotes filters notes by fuzzy matching query against titles.
// Returns all items if query is empty.
func filterNotes(query string, items []notedeprecated.Note) []notedeprecated.Note {
	if query == "" {
		return items
	}

	titles := make([]string, len(items))
	for i, n := range items {
		titles[i] = n.Title
	}

	matches := fuzzy.Find(query, titles)
	result := make([]notedeprecated.Note, len(matches))
	for i, match := range matches {
		result[i] = items[match.Index]
	}

	return result
}

// calcCursorUp/calcCursorDown mirror listpanel's cursor math; the filtered
// result list pages the same way the note list does.
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

func (m *Modal) Filter(query string) {
	m.Filtered = filterNotes(query, m.AllItems)
	m.Cursor = 0
	m.Offset = 0
}

func (m *Modal) CursorUp() {
	m.Cursor, m.Offset = calcCursorUp(m.Cursor, m.Offset)
}

func (m *Modal) CursorDown() {
	m.Cursor, m.Offset = calcCursorDown(m.Cursor, len(m.Filtered), m.Offset, m.Height)
}

func (m Modal) SelectedItem() notedeprecated.Note {
	if len(m.Filtered) == 0 {
		return notedeprecated.Note{}
	}
	return m.Filtered[m.Cursor]
}
