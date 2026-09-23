package tui

import (
	"image"

	"notebox/internal/tui/styles"
)

// listWidthRatio is the note list's share of the screen width (1/4).
const listWidthRatio = 4

// uiLayout is where each part of the screen goes. It's recomputed only when
// the terminal is resized.
type uiLayout struct {
	area    image.Rectangle
	header  image.Rectangle
	list    image.Rectangle
	preview image.Rectangle
	status  image.Rectangle
}

func generateLayout(sty *styles.Styles, width, height int) uiLayout {
	area := image.Rect(0, 0, width, height)
	headerHeight := 1 + sty.Header.GetVerticalFrameSize()
	const statusHeight = 1

	bodyTop := min(headerHeight, height)
	bodyBottom := max(bodyTop, height-statusHeight)
	listRight := width / listWidthRatio

	return uiLayout{
		area:    area,
		header:  image.Rect(0, 0, width, bodyTop),
		list:    image.Rect(0, bodyTop, listRight, bodyBottom),
		preview: image.Rect(listRight, bodyTop, width, bodyBottom),
		status:  image.Rect(0, bodyBottom, width, height),
	}
}
