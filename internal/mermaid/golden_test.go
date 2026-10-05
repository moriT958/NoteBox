package mermaid

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "write the drawings of testdata/**/*.mmd to their .golden files")

// TestGolden draws each diagram in testdata and compares it with the
// drawing in the .golden file next to it.
func TestGolden(t *testing.T) {
	files, err := filepath.Glob("testdata/*/*.mmd")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no diagrams in testdata")
	}
	for _, file := range files {
		name := strings.TrimSuffix(file, ".mmd")
		t.Run(strings.TrimPrefix(name, "testdata/"), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Render(string(src), Styles{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			got += "\n"

			golden := name + ".golden"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run go test -update to write it)", err)
			}
			if got != string(want) {
				t.Errorf("got\n%s\nwant\n%s", got, want)
			}
		})
	}
}
