package diagram

import (
	"strconv"
	"strings"
)

var namedEntities = map[string]string{
	"lt": "<", "gt": ">", "amp": "&", "quot": `"`, "apos": "'", "nbsp": " ",
}

// decodeEntities decodes HTML entities such as &lt; and &#9829;, and
// Mermaid's own form of them such as #quot; and #9829;. Anything that
// doesn't decode is kept as is.
func decodeEntities(s string) string {
	var b strings.Builder
	for {
		i := strings.IndexAny(s, "&#")
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		s = s[i:]
		if text, n := entityAt(s); n > 0 {
			b.WriteString(text)
			s = s[n:]
			continue
		}
		b.WriteByte(s[0])
		s = s[1:]
	}
}

// entityAt decodes the entity that s starts with and reports its length, or
// a length of zero when s doesn't start with one.
func entityAt(s string) (string, int) {
	// Long enough for the longest names and code points.
	const maxBody = 10

	// Mermaid writes both named and numeric entities after a "#", where HTML
	// writes "&" and "&#".
	mermaid := s[0] == '#'
	body := s[1:]
	numericOnly := false
	if !mermaid && strings.HasPrefix(body, "#") {
		body, numericOnly = body[1:], true
	}
	semi := strings.IndexByte(body, ';')
	if semi <= 0 || semi > maxBody {
		return "", 0
	}
	name := body[:semi]
	n := len(s) - len(body) + semi + 1

	if !numericOnly {
		if text, ok := namedEntities[name]; ok {
			return text, n
		}
		if !mermaid {
			return "", 0
		}
	}
	base, digits := 10, name
	if d, ok := strings.CutPrefix(strings.ToLower(name), "x"); ok {
		base, digits = 16, d
	}
	code, err := strconv.ParseUint(digits, base, 32)
	if err != nil || code == 0 || code > 0x10ffff || code >= 0xd800 && code <= 0xdfff {
		return "", 0
	}
	return string(rune(code)), n
}
