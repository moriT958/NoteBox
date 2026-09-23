package note

type Note struct {
	title string
	path  string
}

func (n Note) Title() string {
	return n.title
}

// Path returns the note's file path relative to its box.
func (n Note) Path() string {
	return n.path
}
