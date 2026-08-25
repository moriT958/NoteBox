package note

import "context"

type Note struct {
	Title string
	Path  string
}

// Box represents a note collection backed by one or more directories.
// Path is the box's primary directory (used for new notes and as the
// legacy single-directory identity). Paths holds additional directories
// merged into the box; together with Path they all behave as one box.
type Box struct {
	ID     int
	Title  string
	Path   string
	Paths  []string
	Active bool
}

// AllPaths returns every directory backing the box: the primary Path
// followed by any merged Paths.
func (b Box) AllPaths() []string {
	return append([]string{b.Path}, b.Paths...)
}

type BoxRepository interface {
	FindAll(context.Context) ([]Box, error)
	FindAllActive(context.Context) ([]Box, error)
	FindInactiveBoxes(context.Context) ([]Box, error)
	PruneBoxes(context.Context) error
	CreateBox(context.Context, Box) (Box, error)
	UpdateBox(context.Context, Box) (Box, error)
	DeleteBox(context.Context, Box) error
	AddBoxPath(ctx context.Context, box Box, path string) (Box, error)
	RemoveBoxPath(ctx context.Context, box Box, path string) (Box, error)
}
