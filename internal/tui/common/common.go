package common

import (
	"image"

	"notebox/internal/config"
	"notebox/internal/tui/styles"

	"charm.land/bubbles/v2/textinput"
	"github.com/charmbracelet/x/ansi"
)

type Common struct {
	Config *config.Config
	Styles *styles.Styles
}

// CenterRect returns a width x height rectangle centered within area.
func CenterRect(area image.Rectangle, width, height int) image.Rectangle {
	x := area.Min.X + (area.Dx()-width)/2
	y := area.Min.Y + (area.Dy()-height)/2
	return image.Rect(x, y, x+width, y+height)
}

// NewInput returns a text input that relies on the real terminal cursor,
// which the root positions from the component's Cursor().
func NewInput(placeholder string, charLimit int) textinput.Model {
	in := textinput.New()
	in.Placeholder = placeholder
	in.CharLimit = charLimit
	in.SetVirtualCursor(false)
	return in
}

// InputCursorX returns the input's cursor column using the display width of
// the text before it; textinput derives it from the rune count, which
// undercounts full-width characters.
func InputCursorX(in textinput.Model) int {
	value := []rune(in.Value())
	pos := min(in.Position(), len(value))
	return ansi.StringWidth(in.Prompt) + ansi.StringWidth(string(value[:pos]))
}
