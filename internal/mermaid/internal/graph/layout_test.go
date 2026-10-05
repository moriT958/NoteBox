package graph

import (
	"maps"
	"slices"
	"strings"
	"testing"
)

// testGraph builds a graph that flows down from nodes written alone as "A",
// and edges written as "A>B", with as many ">" as the edge's length, or as
// "A~B" for an invisible edge. Nodes are added as they first appear, with
// their names as labels, and index maps the names to their indexes.
func testGraph(t *testing.T, specs ...string) (g *Graph, index map[string]int) {
	t.Helper()
	g, index = &Graph{}, map[string]int{}
	node := func(name string) int {
		if name == "" {
			t.Fatalf("testGraph: a node without a name in %q", specs)
		}
		i, ok := index[name]
		if !ok {
			i = len(g.Nodes)
			index[name] = i
			g.Nodes = append(g.Nodes, Node{Label: name})
		}
		return i
	}
	for _, spec := range specs {
		if from, to, ok := strings.Cut(spec, "~"); ok {
			g.Edges = append(g.Edges, Edge{From: node(from), To: node(to), Invisible: true, Length: 1})
			continue
		}
		i := strings.IndexByte(spec, '>')
		if i < 0 {
			node(spec)
			continue
		}
		j := i
		for j < len(spec) && spec[j] == '>' {
			j++
		}
		g.Edges = append(g.Edges, Edge{From: node(spec[:i]), To: node(spec[j:]), ToHead: HeadArrow, Length: j - i})
	}
	return g, index
}

func TestRankNodes(t *testing.T) {
	tests := []struct {
		name  string
		edges []string
		want  map[string]int
	}{
		{
			name:  "chain",
			edges: []string{"A>B", "B>C"},
			want:  map[string]int{"A": 0, "B": 1, "C": 2},
		},
		{
			name:  "below the lowest node with an edge to it",
			edges: []string{"A>B", "B>C", "A>C"},
			want:  map[string]int{"A": 0, "B": 1, "C": 2},
		},
		{
			name:  "longer edges",
			edges: []string{"A>>B", "B>>>>C"},
			want:  map[string]int{"A": 0, "B": 2, "C": 6},
		},
		{
			name:  "nodes without incoming edges move down",
			edges: []string{"A>B", "B>C", "C>D", "E>D"},
			want:  map[string]int{"A": 0, "B": 1, "C": 2, "D": 3, "E": 2},
		},
		{
			name:  "the top rank is never left empty",
			edges: []string{"A>X", "P>Q", "Q>R", "R>X"},
			want:  map[string]int{"A": 2, "X": 3, "P": 0, "Q": 1, "R": 2},
		},
		{
			name:  "nodes without edges",
			edges: []string{"A", "B>C"},
			want:  map[string]int{"A": 0, "B": 0, "C": 1},
		},
		{
			name:  "invisible edges",
			edges: []string{"A~B"},
			want:  map[string]int{"A": 0, "B": 1},
		},
		{
			name:  "edges going back up are turned around",
			edges: []string{"A>B", "B>C", "C>A"},
			want:  map[string]int{"A": 0, "B": 1, "C": 2},
		},
		{
			name:  "edges to the node itself are left out",
			edges: []string{"A>A", "A>B"},
			want:  map[string]int{"A": 0, "B": 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, index := testGraph(t, tt.edges...)
			ranks := rankNodes(g, backEdges(g))
			got := map[string]int{}
			for name, i := range index {
				got[name] = ranks[i]
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("ranks = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackEdges(t *testing.T) {
	tests := []struct {
		name  string
		edges []string
		want  []bool
	}{
		{"no cycles", []string{"A>B", "B>C", "A>C"}, []bool{false, false, false}},
		{"two nodes", []string{"A>B", "B>A"}, []bool{false, true}},
		{"longer cycle", []string{"A>B", "B>C", "C>A"}, []bool{false, false, true}},
		{"declared from the middle", []string{"B>C", "C>A", "A>B"}, []bool{false, false, true}},
		{"self-loops", []string{"A>A"}, []bool{false}},
		{"cross edges", []string{"A>B", "A>C", "C>B"}, []bool{false, false, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, _ := testGraph(t, tt.edges...)
			if got := backEdges(g); !slices.Equal(got, tt.want) {
				t.Errorf("backEdges() = %v, want %v", got, tt.want)
			}
		})
	}
}

// rankOrder returns the labels of the nodes in each rank from left to right,
// with "·" for the points edges pass through.
func rankOrder(g *Graph, l *layout) []string {
	var ranks []string
	for _, rank := range l.ranks {
		var ids []string
		for _, v := range rank {
			if l.verts[v].isPoint() {
				ids = append(ids, "·")
			} else {
				ids = append(ids, g.Nodes[l.verts[v].node].Label)
			}
		}
		ranks = append(ranks, strings.Join(ids, " "))
	}
	return ranks
}

func TestLayoutOrder(t *testing.T) {
	tests := []struct {
		name  string
		edges []string
		want  []string
	}{
		{
			name:  "declaration order",
			edges: []string{"A>B", "A>C", "A>D"},
			want:  []string{"A", "B C D"},
		},
		{
			name:  "points of long edges sit among the other ends' neighbors",
			edges: []string{"A>B", "B>C", "A>C", "D>C"},
			want:  []string{"A", "B · D", "C"},
		},
		{
			name:  "points of long edges",
			edges: []string{"A>>B"},
			want:  []string{"A", "·", "B"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, _ := testGraph(t, tt.edges...)
			l, err := layOut(g)
			if err != nil {
				t.Fatalf("layOut() error = %v", err)
			}
			if got := rankOrder(g, l); !slices.Equal(got, tt.want) {
				t.Errorf("ranks = %q, want %q", got, tt.want)
			}
			if c := l.crossings(); c != 0 {
				t.Errorf("crossings = %d, want 0", c)
			}
		})
	}
}

func TestReduceCrossings(t *testing.T) {
	g, index := testGraph(t, "A", "B", "A>C", "C>E", "B>D", "D>F")
	l, err := layOut(g)
	if err != nil {
		t.Fatalf("layOut() error = %v", err)
	}
	// Start from an order with edges crossing between both pairs of ranks.
	l.ranks = [][]int{
		{index["A"], index["B"]},
		{index["D"], index["C"]},
		{index["E"], index["F"]},
	}
	if c := l.crossings(); c != 2 {
		t.Fatalf("crossings before = %d, want 2", c)
	}
	l.reduceCrossings()
	if got, want := rankOrder(g, l), []string{"A B", "C D", "E F"}; !slices.Equal(got, want) {
		t.Errorf("ranks = %q, want %q", got, want)
	}
	if c := l.crossings(); c != 0 {
		t.Errorf("crossings after = %d, want 0", c)
	}
}

func TestLayoutPlace(t *testing.T) {
	g, index := testGraph(t, "A>B", "A>C", "A>D", "B>E", "C>E", "D>E", "A>>E")
	g.Nodes[index["A"]].Label = "Start"
	g.Nodes[index["B"]].Label = "とても長いラベルを持つノードで、折り返しと省略の動きを確かめるためのもの"
	l, err := layOut(g)
	if err != nil {
		t.Fatalf("layOut() error = %v", err)
	}

	for r, rank := range l.ranks {
		for i, v := range rank {
			vert := l.verts[v]
			if vert.cx-l.left(&vert) < 0 || vert.cx+l.right(&vert) >= l.w {
				t.Errorf("vertex %d at columns %d-%d is outside the layout's width %d", v, vert.cx-l.left(&vert), vert.cx+l.right(&vert), l.w)
			}
			if vert.y < l.rankY[r] || vert.y+vert.h > l.rankY[r]+l.rankH[r] {
				t.Errorf("vertex %d at rows %d-%d is outside its rank's rows %d-%d", v, vert.y, vert.y+vert.h, l.rankY[r], l.rankY[r]+l.rankH[r])
			}
			if i > 0 {
				prev := l.verts[rank[i-1]]
				if gap := (vert.cx - l.left(&vert)) - (prev.cx + l.right(&prev) + 1); gap < l.orient.gapX() {
					t.Errorf("vertices %d and %d in rank %d are %d apart, less than %d", rank[i-1], v, r, gap, l.orient.gapX())
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
	b := l.verts[index["B"]]
	if want := []string{"とても長いラベルを持つノ", "ードで、折り返しと省略の", "動きを確かめるためのもの"}; !slices.Equal(b.lines, want) {
		t.Errorf("lines = %q, want %q", b.lines, want)
	}
	if b.w != 24+2*labelPad+2 || b.h != 3+2 {
		t.Errorf("box = %dx%d, want %dx%d", b.w, b.h, 24+2*labelPad+2, 3+2)
	}
}

func TestLayoutCenterUnderSingleParent(t *testing.T) {
	g, index := testGraph(t, "A>B")
	g.Nodes[index["A"]].Label = "a wide parent node"
	l, err := layOut(g)
	if err != nil {
		t.Fatalf("layOut() error = %v", err)
	}
	if a, b := l.verts[index["A"]], l.verts[index["B"]]; a.cx != b.cx {
		t.Errorf("parent centered at %d, child at %d", a.cx, b.cx)
	}
}
