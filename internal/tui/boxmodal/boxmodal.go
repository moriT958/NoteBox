// Package boxmodal owns the box list modal's state: its cached item list
// and the create/rename/delete/load lifecycle operations that keep it in
// sync with the repository.
package boxmodal

import (
	"notebox/internal/note"

	"charm.land/bubbles/v2/textinput"
)

// Mode selects which form the create modal renders.
type Mode int

const (
	ModeNewBox Mode = iota
	ModeOpenFolder
)

// Field identifies which text input in the create modal is focused.
type Field int

const (
	TitleField Field = iota
	PathField
)

type BoxModal struct {
	Width, Height int
	Cursor        int
	Offset        int
	Items         []note.Box
	Mode          Mode
	TitleInput    textinput.Model
	PathInput     textinput.Model
	RenameInput   textinput.Model
	ActiveField   Field
	ValidationErr string
}

// calcCursorUp/calcCursorDown mirror listpanel's cursor math; box lists use
// the same paging behavior as the note list.
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

func (m *BoxModal) CursorUp() {
	m.Cursor, m.Offset = calcCursorUp(m.Cursor, m.Offset)
}

func (m *BoxModal) CursorDown() {
	m.Cursor, m.Offset = calcCursorDown(m.Cursor, len(m.Items), m.Offset, m.Height)
}

func (m BoxModal) SelectedItem() note.Box {
	if len(m.Items) == 0 || m.Cursor >= len(m.Items) {
		return note.Box{}
	}
	return m.Items[m.Cursor]
}

// ApplyBoxesLoaded installs a freshly loaded box list and puts the cursor on
// the currently active box.
func (m *BoxModal) ApplyBoxesLoaded(boxes []note.Box, currentBoxID int) {
	m.Items = boxes
	m.Cursor = 0
	m.Offset = 0
	for i, b := range boxes {
		if b.ID == currentBoxID {
			m.Cursor = i
			break
		}
	}
}

// ApplyBoxCreated appends a newly created box to the cached list.
func (m *BoxModal) ApplyBoxCreated(box note.Box) {
	m.Items = append(m.Items, box)
}

// ApplyBoxDeleted removes a box from the cached list by ID and keeps the
// cursor within bounds.
func (m *BoxModal) ApplyBoxDeleted(id int) {
	for i, b := range m.Items {
		if b.ID == id {
			m.Items = append(m.Items[:i], m.Items[i+1:]...)
			if m.Cursor >= len(m.Items) && m.Cursor > 0 {
				m.Cursor--
			}
			return
		}
	}
}

// ApplyBoxRenamed replaces a box in the cached list with its updated value.
func (m *BoxModal) ApplyBoxRenamed(box note.Box) {
	for i, b := range m.Items {
		if b.ID == box.ID {
			m.Items[i] = box
			return
		}
	}
}
