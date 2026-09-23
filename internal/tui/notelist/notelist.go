package notelist

import (
	"slices"
	"strings"

	"notebox/internal/core/note"
	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// itemLines is the rows each note occupies: title, path and a spacer.
const itemLines = 3

const (
	gutterBar   = "│"
	gutter      = gutterBar + " "
	plainGutter = "  "
)

type NoteList struct {
	com *common.Common

	// width and height include the border.
	width, height int

	notes  []note.Note
	cursor int
	// offset is the index of the first visible note.
	offset int

	renaming bool
	input    textinput.Model
}

func New(com *common.Common) *NoteList {
	return &NoteList{
		com:   com,
		input: common.NewInput("", 100),
	}
}

func (l *NoteList) SetSize(width, height int) {
	l.width, l.height = width, height
	l.input.SetWidth(max(1, l.contentWidth()-ansi.StringWidth(gutter)-1))
	l.clamp()
}

func (l *NoteList) contentWidth() int {
	return max(0, l.width-l.com.Styles.BorderFocused.GetHorizontalFrameSize())
}

func (l *NoteList) contentHeight() int {
	return max(0, l.height-l.com.Styles.BorderFocused.GetVerticalFrameSize())
}

func (l *NoteList) visibleItems() int {
	return max(1, l.contentHeight()/itemLines)
}

// SetNotes replaces the list, keeping the selected note selected if it still
// exists and otherwise keeping the cursor at the same index.
func (l *NoteList) SetNotes(notes []note.Note) {
	selected, ok := l.Selected()
	l.notes = notes
	if ok {
		if i := slices.Index(notes, selected); i >= 0 {
			l.cursor = i
		}
	}
	l.clamp()
}

func (l *NoteList) Notes() []note.Note {
	return l.notes
}

// Add appends n and selects it.
func (l *NoteList) Add(n note.Note) {
	l.notes = append(l.notes, n)
	l.Select(n)
}

// Remove drops n; the cursor stays on the same index, i.e. the next note.
func (l *NoteList) Remove(n note.Note) {
	l.SetNotes(slices.DeleteFunc(slices.Clone(l.notes), func(x note.Note) bool { return x == n }))
}

// Replace swaps old for n in place, e.g. after a rename.
func (l *NoteList) Replace(old, n note.Note) {
	if i := slices.Index(l.notes, old); i >= 0 {
		l.notes[i] = n
	}
}

// Select moves the cursor to n and reports whether n is in the list.
func (l *NoteList) Select(n note.Note) bool {
	i := slices.Index(l.notes, n)
	if i < 0 {
		return false
	}
	l.cursor = i
	l.clamp()
	return true
}

func (l *NoteList) Selected() (note.Note, bool) {
	if l.cursor < 0 || l.cursor >= len(l.notes) {
		return note.Note{}, false
	}
	return l.notes[l.cursor], true
}

func (l *NoteList) CursorUp() {
	l.cursor--
	l.clamp()
}

func (l *NoteList) CursorDown() {
	l.cursor++
	l.clamp()
}

// clamp keeps the cursor inside the list and scrolls it into view.
func (l *NoteList) clamp() {
	l.cursor = max(0, min(l.cursor, len(l.notes)-1))
	visible := l.visibleItems()
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+visible {
		l.offset = l.cursor - visible + 1
	}
	l.offset = max(0, min(l.offset, len(l.notes)-visible))
}

// StartRename opens the inline rename input on the selected note.
func (l *NoteList) StartRename() bool {
	n, ok := l.Selected()
	if !ok {
		return false
	}
	l.renaming = true
	l.input.SetValue(n.Title())
	l.input.CursorEnd()
	l.input.Focus()
	return true
}

func (l *NoteList) Renaming() bool {
	return l.renaming
}

func (l *NoteList) RenameValue() string {
	return l.input.Value()
}

func (l *NoteList) StopRename() {
	l.renaming = false
	l.input.Blur()
}

// UpdateRenameInput forwards key presses and pastes to the rename input.
func (l *NoteList) UpdateRenameInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	l.input, cmd = l.input.Update(msg)
	return cmd
}

// Cursor returns the rename input's cursor relative to the list's top-left
// corner, or nil when not renaming.
func (l *NoteList) Cursor() *tea.Cursor {
	if !l.renaming {
		return nil
	}
	border := l.com.Styles.BorderFocused
	x := border.GetBorderLeftSize() + ansi.StringWidth(gutter) + common.InputCursorX(l.input)
	y := border.GetBorderTopSize() + (l.cursor-l.offset)*itemLines
	return tea.NewCursor(x, y)
}

func (l *NoteList) Render(focused bool) string {
	sty := l.com.Styles
	w := l.contentWidth()

	var b strings.Builder
	if len(l.notes) == 0 {
		b.WriteString("no items.")
	}
	end := min(l.offset+l.visibleItems(), len(l.notes))
	for i := l.offset; i < end; i++ {
		n := l.notes[i]
		if i != l.offset {
			b.WriteString("\n")
		}
		if i != l.cursor {
			b.WriteString(ansi.Truncate(plainGutter+n.Title(), w, "…"))
			b.WriteString("\n")
			b.WriteString(ansi.Truncate(plainGutter+sty.List.Path.Render(n.Path()), w, "…"))
			b.WriteString("\n")
			continue
		}

		g := sty.List.Cursor.Render(gutterBar) + " "
		if l.renaming {
			b.WriteString(g)
			b.WriteString(l.input.View())
		} else {
			b.WriteString(ansi.Truncate(g+sty.List.Cursor.Render(n.Title()), w, "…"))
		}
		b.WriteString("\n")
		b.WriteString(ansi.Truncate(g+sty.List.SelectedPath.Render(n.Path()), w, "…"))
		b.WriteString("\n")
	}

	border := sty.BorderBlurred
	if focused {
		border = sty.BorderFocused
	}
	return border.Width(l.width).Height(l.height).Render(b.String())
}
