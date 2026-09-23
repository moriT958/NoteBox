package note

import (
	"fmt"
	"strconv"
	"strings"
)

const filenameEscapeByte = '%'

func isInvalidFilenameByte(b byte) bool {
	switch b {
	case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
		return true
	}
	return b < 0x20
}

// encodeTitle escapes bytes that are unsafe in a filename as %XX, leaving
// everything else ( including multi-byte UTF-8 runes ) untouched, so that
// decodeTitle can always recover the exact original title.
func encodeTitle(title string) string {
	var b strings.Builder
	for i := 0; i < len(title); i++ {
		c := title[i]
		if c == filenameEscapeByte || isInvalidFilenameByte(c) {
			fmt.Fprintf(&b, "%%%02X", c)
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func decodeTitle(name string) string {
	var b strings.Builder
	for i := 0; i < len(name); i++ {
		if name[i] == filenameEscapeByte && i+2 < len(name) {
			if v, err := strconv.ParseUint(name[i+1:i+3], 16, 8); err == nil {
				b.WriteByte(byte(v))
				i += 2
				continue
			}
		}
		b.WriteByte(name[i])
	}
	return b.String()
}
