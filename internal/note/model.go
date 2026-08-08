package note

import "context"

type Note struct {
	Title string
	Path  string
}

type Box struct {
	ID     int
	Title  string
	Path   string
	Active bool
}

type BoxRepository interface {
	FindAll(context.Context) ([]Box, error)
	FindAllActive(context.Context) ([]Box, error)
	CreateBox(context.Context, Box) (Box, error)
	UpdateBox(context.Context, Box) (Box, error)
	DeleteBox(context.Context, Box) error
}
