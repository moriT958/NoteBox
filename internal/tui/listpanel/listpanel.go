// Package listpanel owns the note list: its items, cursor/offset scrolling
// math, and the selection operations other modals (fuzzy search, box
// switch) go through instead of recomputing cursor positioning themselves.
package listpanel

import (
	"notebox/internal/notedeprecated"
	"slices"

	"charm.land/bubbles/v2/textinput"
)

// NoteItemLines is the number of terminal rows each note item occupies: the
// title, the path below it, and a blank spacer row separating it from the
// next item. Callers use it to convert a content height into a row count.
const NoteItemLines = 3

type ListPanel struct {
	Width, Height int
	// BoxRows is the full available content height in terminal rows. It can
	// exceed Height*NoteItemLines when the content height doesn't divide
	// evenly, and is used to size the border box so it lines up with the
	// previewer panel instead of coming up short.
	BoxRows     int
	Cursor      int
	Items       []notedeprecated.Note
	Offset      int
	RenameInput textinput.Model

	// notes dir change watcher
	Registerer   notedeprecated.Registerer
	NotesUpdates <-chan []notedeprecated.Note
}

// calcCursorUp calculates the new cursor and offset when moving up.
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

// calcCursorDown calculates the new cursor and offset when moving down.
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

// calcRemoveItem calculates the new items and cursor after removing an item.
func calcRemoveItem(items []notedeprecated.Note, cursor int) ([]notedeprecated.Note, int) {
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

// calcAddItem calculates the new cursor and offset after adding an item.
func calcAddItem(itemCount, offset, height int) (newCursor, newOffset int) {
	newCursor = itemCount - 1
	newOffset = offset
	if newCursor >= offset+height {
		newOffset = newCursor - height + 1
	}
	return
}

// preserveSelectionPos calculates cursor and offset after notes reload while
// keeping the cursor visible.
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

func (p *ListPanel) CursorUp() {
	p.Cursor, p.Offset = calcCursorUp(p.Cursor, p.Offset)
}

func (p *ListPanel) CursorDown() {
	p.Cursor, p.Offset = calcCursorDown(p.Cursor, len(p.Items), p.Offset, p.Height)
}

// SelectByIndex jumps the cursor to item index i and repositions the
// viewport window to it. This is the "select this exact item" seam other
// modals (fuzzy search, box switch) go through instead of recomputing
// cursor/offset positioning themselves.
func (p *ListPanel) SelectByIndex(i int) {
	if i < 0 || i >= len(p.Items) {
		return
	}
	p.Cursor = i
	if p.Cursor >= p.Height {
		p.Offset = p.Cursor - p.Height + 1
	} else {
		p.Offset = 0
	}
}

// SelectedItem returns the item under the cursor.
func (p ListPanel) SelectedItem() notedeprecated.Note {
	if p.Cursor < 0 || len(p.Items) == 0 || len(p.Items) <= p.Cursor {
		return notedeprecated.Note{}
	}
	return p.Items[p.Cursor]
}

// AddItem appends n and adjusts cursor/offset to keep it visible.
func (p *ListPanel) AddItem(n notedeprecated.Note) {
	p.Items = append(p.Items, n)
	p.Cursor, p.Offset = calcAddItem(len(p.Items), p.Offset, p.Height)
}

// RemoveItem removes the item under the cursor.
func (p *ListPanel) RemoveItem() {
	p.Items, p.Cursor = calcRemoveItem(p.Items, p.Cursor)
}

// ReloadAllNotes replaces the item list (e.g. after an fsnotify change)
// while keeping the previously selected note (by path) selected when it
// still exists, and clamping cursor/offset to the new item count.
func (p *ListPanel) ReloadAllNotes(notes []notedeprecated.Note) {
	selectedPath := p.SelectedItem().Path
	p.Items = notes

	for i, n := range notes {
		if n.Path == selectedPath {
			p.Cursor = i
			break
		}
	}

	p.Cursor, p.Offset = preserveSelectionPos(p.Cursor, p.Offset, p.Height, len(notes))
}

// SetSize resizes the panel and re-clamps cursor/offset to fit.
func (p *ListPanel) SetSize(width, boxRows, height int) {
	p.Width = width
	p.BoxRows = boxRows
	p.Height = height
	p.RenameInput.SetWidth(width - 4)
	p.Cursor, p.Offset = preserveSelectionPos(p.Cursor, p.Offset, p.Height, len(p.Items))
}

// Reset points the panel at a freshly registered notes directory, clearing
// the item list and cursor/offset back to the top.
func (p *ListPanel) Reset(notesUpdates <-chan []notedeprecated.Note) {
	p.NotesUpdates = notesUpdates
	p.Items = nil
	p.Cursor = 0
	p.Offset = 0
}
