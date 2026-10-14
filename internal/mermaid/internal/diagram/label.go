package diagram

import "strings"

// formatTags are the HTML tags that only style text. Labels drop them and
// keep their text.
var formatTags = map[string]bool{
	"b": true, "strong": true, "i": true, "em": true, "u": true, "s": true,
	"strike": true, "del": true, "ins": true, "mark": true, "small": true,
	"big": true, "sub": true, "sup": true, "code": true, "kbd": true,
	"samp": true, "var": true, "tt": true, "span": true, "font": true,
	"q": true, "abbr": true, "cite": true, "pre": true,
}

// CleanLabel turns the text of a label as written in the source into the
// text to draw. It removes the quotes around it, the markers of a Markdown
// string and formatting tags, turns <br> into a line break and decodes
// entities. Runs of white space within a line become a single space.
func CleanLabel(raw string) string {
	s := strings.TrimSpace(raw)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		s = stripMarkdown(s[1 : len(s)-1])
	}
	// Tags go before entities, so that an escaped tag such as &lt;b&gt;
	// stays as text.
	s = decodeEntities(stripTags(s))

	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.Join(strings.Fields(l), " ")
	}
	return strings.Join(lines, "\n")
}

// stripMarkdown removes the emphasis markers of a Markdown string, leaving
// underscores and asterisks inside words, as in snake_case.
func stripMarkdown(s string) string {
	s = strings.NewReplacer("**", "", "__", "", "`", "").Replace(s)
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		if r == '*' || r == '_' {
			inWord := i > 0 && i < len(rs)-1 && isWordRune(rs[i-1]) && isWordRune(rs[i+1])
			if !inWord {
				continue
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isWordRune(r rune) bool {
	return r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= 0x80
}

// stripTags removes formatting tags and turns <br> into a line break. Other
// tags are kept as text.
func stripTags(s string) string {
	var b strings.Builder
	for {
		i := strings.IndexByte(s, '<')
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		s = s[i:]
		name, n := tagAt(s)
		switch {
		case n > 0 && name == "br":
			b.WriteByte('\n')
		case n > 0 && formatTags[name]:
		default:
			b.WriteByte('<')
			n = 1
		}
		s = s[n:]
	}
}

// tagAt reports the lowercased name of the tag that s starts with, such as
// "br" for "<br/>" or "b" for "</b>", and its length. It returns a length of
// zero when s doesn't start with a tag.
func tagAt(s string) (string, int) {
	end := strings.IndexByte(s, '>')
	if end < 0 || strings.IndexByte(s[1:end], '<') >= 0 {
		return "", 0
	}
	inner := strings.TrimPrefix(s[1:end], "/")
	j := 0
	for j < len(inner) && isASCIIAlnum(inner[j]) {
		j++
	}
	if j == 0 {
		return "", 0
	}
	return strings.ToLower(inner[:j]), end + 1
}

func isASCIIAlnum(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
