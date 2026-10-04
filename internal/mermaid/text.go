package mermaid

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// labelBreakpoints are where words too long for a line may break, in
// addition to hyphens, so identifiers split between their segments.
const labelBreakpoints = "_./"

// formatTags are the HTML tags that only style text. Labels drop them and
// keep their text.
var formatTags = map[string]bool{
	"b": true, "strong": true, "i": true, "em": true, "u": true, "s": true,
	"strike": true, "del": true, "ins": true, "mark": true, "small": true,
	"big": true, "sub": true, "sup": true, "code": true, "kbd": true,
	"samp": true, "var": true, "tt": true, "span": true, "font": true,
	"q": true, "abbr": true, "cite": true, "pre": true,
}

var namedEntities = map[string]string{
	"lt": "<", "gt": ">", "amp": "&", "quot": `"`, "apos": "'", "nbsp": " ",
}

// cleanLabel turns the text of a label as written in the source into the
// text to draw. It removes the quotes around it, the markers of a Markdown
// string and formatting tags, turns <br> into a line break and decodes
// entities. Runs of white space within a line become a single space.
func cleanLabel(raw string) string {
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

// wrapLabel breaks a cleaned label into lines no wider than width. A label
// that needs more than maxLines lines is cut short with an ellipsis.
func wrapLabel(label string, width, maxLines int) []string {
	var lines []string
	for para := range strings.SplitSeq(label, "\n") {
		for l := range strings.SplitSeq(ansi.Wrap(para, width, labelBreakpoints), "\n") {
			lines = append(lines, strings.TrimSpace(l))
		}
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := lines[maxLines-1]
		if ansi.StringWidth(last) >= width {
			last = ansi.Truncate(last, width-1, "")
		}
		lines[maxLines-1] = last + "…"
	}
	return lines
}

// fitLabel cuts a single-line label short with an ellipsis if it is wider
// than width.
func fitLabel(label string, width int) string {
	return ansi.Truncate(label, width, "…")
}

// textWidth is the number of cells s takes up in a terminal.
func textWidth(s string) int {
	return ansi.StringWidth(s)
}
