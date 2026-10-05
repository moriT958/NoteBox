package flowchart

import "strings"

// A statement of a flowchart, and the line of the source it is on.
type statement struct {
	line int
	text string
}

// splitStatements splits a flowchart into statements, which end at line
// breaks and semicolons, dropping front matter, comments, accessibility
// titles and descriptions, and empty statements.
func splitStatements(src string) []statement {
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

	var stmts []statement
	for ; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if isAccessibility(line, "accTitle", ":") || isAccessibility(line, "accDescr", ":") {
			continue
		}
		if isAccessibility(line, "accDescr", "{") {
			for ; i < len(lines) && !strings.Contains(lines[i], "}"); i++ {
			}
			continue
		}
		var cur strings.Builder
		flush := func() {
			if s := strings.TrimSpace(cur.String()); s != "" {
				stmts = append(stmts, statement{line: i + 1, text: s})
			}
			cur.Reset()
		}
		inQuotes := false
	scan:
		for j := 0; j < len(line); j++ {
			c := line[j]
			switch {
			case c == '"':
				inQuotes = !inQuotes
			case inQuotes:
			case c == '%' && strings.HasPrefix(line[j:], "%%"):
				break scan
			case c == ';':
				flush()
				continue
			}
			cur.WriteByte(c)
		}
		flush()
	}
	return stmts
}

// isAccessibility reports whether line is an accessibility statement that
// starts with keyword and then sep, which runs to the end of the line, or,
// for a "{", to the closing "}".
func isAccessibility(line, keyword, sep string) bool {
	rest, ok := strings.CutPrefix(line, keyword)
	return ok && strings.HasPrefix(strings.TrimSpace(rest), sep)
}
