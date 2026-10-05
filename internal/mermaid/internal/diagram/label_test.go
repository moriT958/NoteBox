package diagram

import (
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
			if got := CleanLabel(tt.in); got != tt.want {
				t.Errorf("CleanLabel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
