package box

type Box struct {
	id     int
	title  string
	path   string
	active bool
}

func (b Box) Path() string {
	return b.path
}

func (b Box) IsActive() bool {
	return b.active
}
