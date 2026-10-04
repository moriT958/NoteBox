package mermaid

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTooLarge(t *testing.T) {
	// A wide fan-out and a long chain, within the limits on edges.
	var wideAndTall strings.Builder
	wideAndTall.WriteString("graph TD\n")
	for i := range 250 {
		fmt.Fprintf(&wideAndTall, "A --> W%d[a label as wide as labels go]\n", i)
	}
	for i := range 249 {
		fmt.Fprintf(&wideAndTall, "C%d --> C%d\n", i, i+1)
	}

	var manyEdges strings.Builder
	manyEdges.WriteString("graph TD\n")
	for i := range maxEdges + 1 {
		fmt.Fprintf(&manyEdges, "A --> N%d\n", i)
	}

	tests := []struct {
		name string
		src  string
	}{
		{"long source", "graph TD\n" + strings.Repeat("%% comment\n", maxTextSize/10)},
		{"many edges", manyEdges.String()},
		{"many edges from groups", "graph TD\n" + strings.Repeat("A & ", 30) + "A --> " + strings.Repeat("B & ", 30) + "B"},
		{"long edges", "graph TD\nA " + strings.Repeat("-", maxPoints+3) + "> B"},
		{"large drawing", wideAndTall.String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Render(tt.src, Styles{}); !errors.Is(err, ErrTooLarge) {
				t.Errorf("Render() error = %v, want ErrTooLarge", err)
			}
		})
	}
}

func TestRenderWithinLimits(t *testing.T) {
	var src strings.Builder
	src.WriteString("graph TD\n")
	for i := range maxEdges {
		fmt.Fprintf(&src, "N%d --> N%d\n", i%50, 50+i%40)
	}
	if _, err := Render(src.String(), Styles{}); err != nil {
		t.Errorf("Render() error = %v", err)
	}
}

func FuzzRender(f *testing.F) {
	files, err := filepath.Glob("testdata/*/*.mmd")
	if err != nil {
		f.Fatal(err)
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(src))
	}
	for _, src := range []string{
		"graph TD\nA-->B",
		"graph LR\nA -- text --> B -.-> C ==> D ~~~ E",
		"flowchart RL\nA[\"a ] b\"] --o B((c)) --x C{d}\nC --> A\nB --> B",
		"graph BT;A & B --> C & D;C -->|x| A",
		"graph TD\nA[\"`**md**`\"] <--> B[ノート<br>作成 &lt;#9829;]",
	} {
		f.Add(src)
	}

	f.Fuzz(func(t *testing.T, src string) {
		out, err := Render(src, Styles{})
		switch {
		case errors.Is(err, errPanic):
			t.Fatalf("Render(%q) panicked: %v", src, err)
		case err != nil:
			if !errors.Is(err, ErrUnsupported) && !errors.Is(err, ErrSyntax) && !errors.Is(err, ErrTooLarge) {
				t.Fatalf("Render(%q) error = %v, want one of the errors of the package", src, err)
			}
		case strings.ContainsAny(out, "\x1b\x00"):
			t.Fatalf("Render(%q) wrote control characters: %q", src, out)
		}
	})
}
