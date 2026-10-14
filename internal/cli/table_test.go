package cli

import (
	"bytes"
	"testing"
)

func TestWriteTable(t *testing.T) {
	header := []string{"Box", "Path"}
	rows := [][]string{{"Default", "/a"}, {"日本語", "/b"}}

	t.Run("Print tab-separated lines without header, when not a terminal.", func(t *testing.T) {
		var buf bytes.Buffer
		writeTable(&buf, false, header, rows, []bool{false, true})

		want := "Default\t/a\n日本語\t/b\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})

	t.Run("Print an aligned table with header and dimmed rows, on a terminal.", func(t *testing.T) {
		var buf bytes.Buffer
		writeTable(&buf, true, header, rows, []bool{false, true})

		want := ansiHeader + "Box      Path" + ansiReset + "\n" +
			"Default  /a\n" +
			ansiDim + "日本語   /b" + ansiReset + "\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})

	t.Run("Print only the header, when there are no rows on a terminal.", func(t *testing.T) {
		var buf bytes.Buffer
		writeTable(&buf, true, header, nil, nil)

		want := ansiHeader + "Box  Path" + ansiReset + "\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})
}
