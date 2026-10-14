package mermaid

import "strings"

// diagramKeyword returns the first word of a diagram's declaration, such as
// "flowchart" in "flowchart LR", skipping the front matter, comments and
// blank lines that may come before it.
func diagramKeyword(src string) string {
	lines := strings.Split(src, "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "---"; i++ {
		}
		i++
	}
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		word, _, _ := strings.Cut(strings.Fields(line)[0], ";")
		return word
	}
	return ""
}
