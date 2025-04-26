package listpanel

import "os"

type Note struct {
	title string
	desc  string
	path  string
}

/* Implement list.Item */
func (n Note) FilterValue() string { return n.title }

/* Implement list.DefaultItem */
func (n Note) Title() string       { return n.title }
func (n Note) Description() string { return n.desc }

func (n Note) Content() (string, error) {
	content, err := os.ReadFile(n.path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
