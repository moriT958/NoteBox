package box

import (
	"fmt"
	"strings"
)

const dirNameEscapeByte = '%'

func isInvalidDirNameByte(b byte) bool {
	switch b {
	case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
		return true
	}
	return b < 0x20
}

// encodeDirName turns a box title into a filesystem-safe directory name.
// Unlike note.encodeTitle this doesn't need to be reversible, since a
// Box's title is stored independently of its path.
func encodeDirName(title string) string {
	var b strings.Builder
	for i := 0; i < len(title); i++ {
		c := title[i]
		if c == dirNameEscapeByte || isInvalidDirNameByte(c) {
			fmt.Fprintf(&b, "%%%02X", c)
			continue
		}
		b.WriteByte(c)
	}
	if b.Len() == 0 {
		return "untitled"
	}
	return b.String()
}
