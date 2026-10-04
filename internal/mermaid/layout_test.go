package mermaid

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

func mustParseFlowchart(t *testing.T, src string) *flowchart {
	t.Helper()
	fc, err := parseFlowchart(src)
	if err != nil {
		t.Fatalf("parseFlowchart(%q) error = %v", src, err)
	}
	return fc
}

func TestRankNodes(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want map[string]int
	}{
		{
			name: "chain",
			src:  "graph TD\nA --> B --> C",
			want: map[string]int{"A": 0, "B": 1, "C": 2},
		},
		{
			name: "below the lowest node with an edge to it",
			src:  "graph TD\nA --> B --> C\nA --> C",
			want: map[string]int{"A": 0, "B": 1, "C": 2},
		},
		{
			name: "longer edges",
			src:  "graph TD\nA ---> B\nB -....-> C",
			want: map[string]int{"A": 0, "B": 2, "C": 6},
		},
		{
			name: "nodes without incoming edges move down",
			src:  "graph TD\nA --> B --> C --> D\nE --> D",
			want: map[string]int{"A": 0, "B": 1, "C": 2, "D": 3, "E": 2},
		},
		{
			name: "the top rank is never left empty",
			src:  "graph TD\nA --> X\nP --> Q --> R --> X",
			want: map[string]int{"A": 2, "X": 3, "P": 0, "Q": 1, "R": 2},
		},
		{
			name: "nodes without edges",
			src:  "graph TD\nA\nB --> C",
			want: map[string]int{"A": 0, "B": 0, "C": 1},
		},
		{
			name: "invisible edges",
			src:  "graph TD\nA ~~~ B",
			want: map[string]int{"A": 0, "B": 1},
		},
		{
			name: "edges going back up are turned around",
			src:  "graph TD\nA --> B --> C --> A",
			want: map[string]int{"A": 0, "B": 1, "C": 2},
		},
		{
			name: "edges to the node itself are left out",
			src:  "graph TD\nA --> A --> B",
			want: map[string]int{"A": 0, "B": 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := mustParseFlowchart(t, tt.src)
			ranks := rankNodes(fc, backEdges(fc))
			got := map[string]int{}
			for i, n := range fc.nodes {
				got[n.id] = ranks[i]
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("ranks = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackEdges(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []bool
	}{
		{"no cycles", "graph TD\nA --> B --> C\nA --> C", []bool{false, false, false}},
		{"two nodes", "graph TD\nA --> B\nB --> A", []bool{false, true}},
		{"longer cycle", "graph TD\nA --> B --> C --> A", []bool{false, false, true}},
		{"declared from the middle", "graph TD\nB --> C\nC --> A\nA --> B", []bool{false, false, true}},
		{"self-loops", "graph TD\nA --> A", []bool{false}},
		{"cross edges", "graph TD\nA --> B\nA --> C\nC --> B", []bool{false, false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := backEdges(mustParseFlowchart(t, tt.src)); !slices.Equal(got, tt.want) {
				t.Errorf("backEdges() = %v, want %v", got, tt.want)
			}
		})
	}
}

// rankOrder returns the IDs of the nodes in each rank from left to right,
// with "·" for the points edges pass through.
func rankOrder(fc *flowchart, l *layout) []string {
	var ranks []string
	for _, rank := range l.ranks {
		var ids []string
		for _, v := range rank {
			if l.verts[v].isPoint() {
				ids = append(ids, "·")
			} else {
				ids = append(ids, fc.nodes[l.verts[v].node].id)
			}
		}
		ranks = append(ranks, strings.Join(ids, " "))
	}
	return ranks
}

func TestLayoutOrder(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "declaration order",
			src:  "graph TD\nA --> B\nA --> C\nA --> D",
			want: []string{"A", "B C D"},
		},
		{
			name: "points of long edges sit among the other ends' neighbors",
			src:  "graph TD\nA --> B --> C\nA --> C\nD --> C",
			want: []string{"A", "B · D", "C"},
		},
		{
			name: "points of long edges",
			src:  "graph TD\nA ---> B",
			want: []string{"A", "·", "B"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := mustParseFlowchart(t, tt.src)
			l, err := layOut(fc)
			if err != nil {
				t.Fatalf("layOut() error = %v", err)
			}
			if got := rankOrder(fc, l); !slices.Equal(got, tt.want) {
				t.Errorf("ranks = %q, want %q", got, tt.want)
			}
			if c := l.crossings(); c != 0 {
				t.Errorf("crossings = %d, want 0", c)
			}
		})
	}
}

func TestReduceCrossings(t *testing.T) {
	fc := mustParseFlowchart(t, "graph TD\nA & B\nA --> C --> E\nB --> D --> F")
	l, err := layOut(fc)
	if err != nil {
		t.Fatalf("layOut() error = %v", err)
	}
	// Start from an order with edges crossing between both pairs of ranks.
	l.ranks = [][]int{
		{fc.index["A"], fc.index["B"]},
		{fc.index["D"], fc.index["C"]},
		{fc.index["E"], fc.index["F"]},
	}
	if c := l.crossings(); c != 2 {
		t.Fatalf("crossings before = %d, want 2", c)
	}
	l.reduceCrossings()
	if got, want := rankOrder(fc, l), []string{"A B", "C D", "E F"}; !slices.Equal(got, want) {
		t.Errorf("ranks = %q, want %q", got, want)
	}
	if c := l.crossings(); c != 0 {
		t.Errorf("crossings after = %d, want 0", c)
	}
}

func TestLayoutPlace(t *testing.T) {
	fc := mustParseFlowchart(t, strings.Join([]string{
		"graph TD",
		"A[Start] --> B[とても長いラベルを持つノードで、折り返しと省略の動きを確かめるためのもの]",
		"A --> C & D",
		"B & C & D --> E",
		"A ---> E",
	}, "\n"))
	l, err := layOut(fc)
	if err != nil {
		t.Fatalf("layOut() error = %v", err)
	}

	for r, rank := range l.ranks {
		for i, v := range rank {
			vert := l.verts[v]
			if vert.cx-vert.left() < 0 || vert.cx+vert.right() >= l.w {
				t.Errorf("vertex %d at columns %d-%d is outside the layout's width %d", v, vert.cx-vert.left(), vert.cx+vert.right(), l.w)
			}
			if vert.y < l.rankY[r] || vert.y+vert.h > l.rankY[r]+l.rankH[r] {
				t.Errorf("vertex %d at rows %d-%d is outside its rank's rows %d-%d", v, vert.y, vert.y+vert.h, l.rankY[r], l.rankY[r]+l.rankH[r])
			}
			if i > 0 {
				prev := l.verts[rank[i-1]]
				if gap := (vert.cx - vert.left()) - (prev.cx + prev.right() + 1); gap < gapX {
					t.Errorf("vertices %d and %d in rank %d are %d apart, less than %d", rank[i-1], v, r, gap, gapX)
				}
			}
		}
		if r > 0 {
			if gap := l.rankY[r] - (l.rankY[r-1] + l.rankH[r-1]); gap < gapY {
				t.Errorf("ranks %d and %d are %d apart, less than %d", r-1, r, gap, gapY)
			}
		}
	}
	if last := len(l.ranks) - 1; l.h != l.rankY[last]+l.rankH[last] {
		t.Errorf("height = %d, want the bottom of the last rank, %d", l.h, l.rankY[last]+l.rankH[last])
	}

	// The long label wraps into a box as wide as its widest line.
	b := l.verts[fc.index["B"]]
	if want := []string{"とても長いラベルを持つノ", "ードで、折り返しと省略の", "動きを確かめるためのもの"}; !slices.Equal(b.lines, want) {
		t.Errorf("lines = %q, want %q", b.lines, want)
	}
	if b.w != 24+2*labelPad+2 || b.h != 3+2 {
		t.Errorf("box = %dx%d, want %dx%d", b.w, b.h, 24+2*labelPad+2, 3+2)
	}
}

func TestLayoutCenterUnderSingleParent(t *testing.T) {
	fc := mustParseFlowchart(t, "graph TD\nA[a wide parent node] --> B")
	l, err := layOut(fc)
	if err != nil {
		t.Fatalf("layOut() error = %v", err)
	}
	if a, b := l.verts[fc.index["A"]], l.verts[fc.index["B"]]; a.cx != b.cx {
		t.Errorf("parent centered at %d, child at %d", a.cx, b.cx)
	}
}
