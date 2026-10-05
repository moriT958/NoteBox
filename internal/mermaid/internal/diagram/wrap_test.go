package diagram

import (
	"slices"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWrapLabel(t *testing.T) {
	tests := []struct {
		name  string
		label string
		width int
		lines int
		want  []string
	}{
		{"fits", "short", 12, 4, []string{"short"}},
		{"empty", "", 12, 4, []string{""}},
		{"words", "short words that need wrapping", 12, 4, []string{"short words", "that need", "wrapping"}},
		{"wide characters", "ノートを保存してデータベースに書き込む", 12, 4, []string{"ノートを保存", "してデータベ", "ースに書き込", "む"}},
		{"identifier", "internal/tui/preview", 12, 4, []string{"internal/", "tui/preview"}},
		{"line breaks", "a\nb", 12, 4, []string{"a", "b"}},
		{"too many lines", "one two three four five", 5, 3, []string{"one", "two", "thre…"}},
		{"too many lines, short last line", "a\nb\nc", 5, 2, []string{"a", "b…"}},
		{"too many lines, wide last line", "ノートを保存", 4, 2, []string{"ノー", "ト…"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapLabel(tt.label, tt.width, tt.lines)
			if !slices.Equal(got, tt.want) {
				t.Errorf("WrapLabel(%q, %d, %d) = %q, want %q", tt.label, tt.width, tt.lines, got, tt.want)
			}
			for _, l := range got {
				if w := ansi.StringWidth(l); w > tt.width {
					t.Errorf("line %q is %d wide, more than %d", l, w, tt.width)
				}
			}
		})
	}
}

func TestFitLabel(t *testing.T) {
	tests := []struct {
		label string
		width int
		want  string
	}{
		{"short", 8, "short"},
		{"longer label", 8, "longer …"},
		{"ノートを保存", 7, "ノート…"},
	}
	for _, tt := range tests {
		if got := FitLabel(tt.label, tt.width); got != tt.want {
			t.Errorf("FitLabel(%q, %d) = %q, want %q", tt.label, tt.width, got, tt.want)
		}
	}
}
