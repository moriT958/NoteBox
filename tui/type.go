package tui

import "notebox/models"

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
