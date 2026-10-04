package mermaid

import (
	"slices"
	"testing"
)

func TestCleanLabel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "  ノート作成  ", "ノート作成"},
		{"quoted", `" a (b) "`, "a (b)"},
		{"white space", "a \t  b", "a b"},
		{"markdown string", "`**太字** と *斜体* と snake_case`", "太字 と 斜体 と snake_case"},
		{"quoted markdown string", "\"`__a__`\"", "a"},
		{"line breaks", "a<br>b<br/>c<BR />d", "a\nb\nc\nd"},
		{"formatting tags", "<b>bold</b> <i>it</i> <span class=x>s</span>", "bold it s"},
		{"other tags stay", "<script>x</script> a<b", "<script>x</script> a<b"},
		{"html entities", "&lt;b&gt; &amp; &quot;x&quot; &#9829; &#x2665;", `<b> & "x" ♥ ♥`},
		{"mermaid entities", "#quot;x#quot; #9829; #35;", `"x" ♥ #`},
		{"escaped tags stay text", "&lt;br&gt;", "<br>"},
		{"not entities", "C# & #tag; &unknown; &#xzz; #;", "C# & #tag; &unknown; &#xzz; #;"},
		{"invalid code points", "&#0; &#xd800; &#1114112;", "&#0; &#xd800; &#1114112;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanLabel(tt.in); got != tt.want {
				t.Errorf("cleanLabel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

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
			got := wrapLabel(tt.label, tt.width, tt.lines)
			if !slices.Equal(got, tt.want) {
				t.Errorf("wrapLabel(%q, %d, %d) = %q, want %q", tt.label, tt.width, tt.lines, got, tt.want)
			}
			for _, l := range got {
				if w := textWidth(l); w > tt.width {
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
		if got := fitLabel(tt.label, tt.width); got != tt.want {
			t.Errorf("fitLabel(%q, %d) = %q, want %q", tt.label, tt.width, got, tt.want)
		}
	}
}
