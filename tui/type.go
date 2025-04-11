package tui

import (
	"notebox/models"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
)

type model struct {
	focus     focusedArea
	mode      mode
	warnModal warnModal
	listPanel listPanel
	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	input     textinput.Model
}

type warnModal struct {
	open    bool
	message string
}

type listPanel struct {
	height int
	width  int
	cursor int
	notes  []*models.Note
}
