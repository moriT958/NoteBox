package graph

import (
	"errors"
	"maps"
	"testing"

	"notebox/internal/mermaid/internal/diagram"
)

func TestDrawInvalidGroups(t *testing.T) {
	tests := []struct {
		groups []Group
		edges  []Edge
	}{
		{groups: []Group{{Nodes: []int{2}}}},
		{groups: []Group{{Nodes: []int{0}}, {Nodes: []int{0}}}},
		{groups: []Group{{Groups: []int{1}}, {Groups: []int{0}}}},
		{groups: []Group{{Groups: []int{0}}}},
		{groups: []Group{{Groups: []int{1}}, {}, {Groups: []int{1}}}},
		{groups: []Group{{Nodes: []int{0}}}, edges: []Edge{{From: 1, To: 1, ToGroup: true, Length: 1}}},
		{groups: []Group{{}}, edges: []Edge{{From: 0, To: 1, FromGroup: true, Length: 1}}},
	}
	for _, tt := range tests {
		g := &Graph{Nodes: []Node{{Label: "A"}, {Label: "B"}}, Groups: tt.groups, Edges: tt.edges}
		_, err := Draw(g)
		if err == nil || errors.Is(err, diagram.ErrSyntax) || errors.Is(err, diagram.ErrTooLarge) || errors.Is(err, diagram.ErrUnsupported) {
			t.Errorf("Draw() with groups %+v and edges %+v error = %v, want an error of its own", tt.groups, tt.edges, err)
		}
	}
}

func TestRankGroupEdges(t *testing.T) {
	tests := []struct {
		name string
		// edges are as for testGraph, and S is the group of a and b, with
		// edges to it added after them.
		edges      []string
		groupEdges []Edge
		want       map[string]int
		// back reports whether the first edge to S goes back up, and ends
		// whether it ends at S's frame, from and to.
		back bool
		ends [2]bool
	}{
		{
			name:       "the group below",
			edges:      []string{"X", "a>b"},
			groupEdges: []Edge{{From: 0, To: 0, ToGroup: true}},
			want:       map[string]int{"X": 0, "a": 1, "b": 2},
			ends:       [2]bool{false, true},
		},
		{
			name:       "below the group",
			edges:      []string{"a>b", "X", "Y>X"},
			groupEdges: []Edge{{From: 0, FromGroup: true, To: 2}},
			want:       map[string]int{"a": 0, "b": 1, "Y": 1, "X": 2},
			ends:       [2]bool{true, false},
		},
		{
			name:       "above the group where below makes a cycle",
			edges:      []string{"a>b", "X>b"},
			groupEdges: []Edge{{From: 0, FromGroup: true, To: 2}},
			want:       map[string]int{"X": 0, "a": 1, "b": 2},
			back:       true,
			ends:       [2]bool{true, false},
		},
		{
			name:       "to the first node where both make cycles",
			edges:      []string{"a>X", "X>b"},
			groupEdges: []Edge{{From: 0, FromGroup: true, To: 1}},
			want:       map[string]int{"a": 0, "X": 1, "b": 2},
			ends:       [2]bool{false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, index := testGraph(t, tt.edges...)
			g.Groups = []Group{{Nodes: groupNodes(index, "a", "b")}}
			first := len(g.Edges)
			for _, e := range tt.groupEdges {
				e.Length = 1
				g.Edges = append(g.Edges, e)
			}
			if err := check(g); err != nil {
				t.Fatal(err)
			}
			l, ranks := rankGraph(g)
			got := map[string]int{}
			for name, i := range index {
				got[name] = ranks[i]
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("ranks = %v, want %v", got, tt.want)
			}
			if l.back[first] != tt.back {
				t.Errorf("back = %v, want %v", l.back[first], tt.back)
			}
			if ends := [2]bool{l.ends[first][0].isFrame(), l.ends[first][1].isFrame()}; ends != tt.ends {
				t.Errorf("ends at frames = %v, want %v", ends, tt.ends)
			}
		})
	}
}

// groupNodes returns the indexes of the named nodes.
func groupNodes(index map[string]int, names ...string) []int {
	var ns []int
	for _, n := range names {
		ns = append(ns, index[n])
	}
	return ns
}

func TestLayoutFrames(t *testing.T) {
	tests := []struct {
		name   string
		edges  []string
		groups func(index map[string]int) []Group
		// groupEdges are added after edges, with a length of one.
		groupEdges []Edge
	}{
		{
			name:  "side by side",
			edges: []string{"c1>a2", "a1>a2", "b1>b2", "c1>c2"},
			groups: func(ix map[string]int) []Group {
				return []Group{
					{Label: "one", Nodes: groupNodes(ix, "a1", "a2")},
					{Label: "two", Nodes: groupNodes(ix, "b1", "b2")},
					{Label: "three", Nodes: groupNodes(ix, "c1", "c2")},
				}
			},
		},
		{
			name:  "nested, with edges in and out",
			edges: []string{"A>B", "B>C", "X>B", "C>Y", "A>>>Y", "Y>A"},
			groups: func(ix map[string]int) []Group {
				return []Group{
					{Label: "outer", Nodes: groupNodes(ix, "A"), Groups: []int{1}},
					{Label: "inner", Nodes: groupNodes(ix, "B", "C")},
				}
			},
		},
		{
			name:  "ranks without nodes of the group",
			edges: []string{"A>>>B", "X>Y", "Y>Z", "Z>W"},
			groups: func(ix map[string]int) []Group {
				return []Group{
					{Label: "a long title for a group", Nodes: groupNodes(ix, "A", "B")},
					{Label: "z", Nodes: groupNodes(ix, "Z")},
				}
			},
		},
		{
			name:  "edges to groups",
			edges: []string{"X>A", "A>B", "B>C", "Y", "Z"},
			groups: func(ix map[string]int) []Group {
				return []Group{
					{Label: "outer", Nodes: groupNodes(ix, "A"), Groups: []int{1}},
					{Label: "inner", Nodes: groupNodes(ix, "B", "C")},
				}
			},
			groupEdges: []Edge{
				{From: 0, To: 0, ToGroup: true, ToHead: HeadArrow, Label: "into"},
				{From: 1, FromGroup: true, To: 4, ToHead: HeadArrow},
				{From: 0, FromGroup: true, To: 5, FromHead: HeadArrow, ToHead: HeadArrow},
				{From: 5, To: 0, ToGroup: true, ToHead: HeadArrow, Label: "back"},
				{From: 1, FromGroup: true, To: 1, ToGroup: true},
			},
		},
		{
			name:  "groups without nodes",
			edges: []string{"A>B"},
			groups: func(ix map[string]int) []Group {
				return []Group{{Label: "empty", Groups: []int{1}}, {Label: "also empty"}, {Label: "b", Nodes: groupNodes(ix, "B")}}
			},
		},
	}
	for _, tt := range tests {
		for _, dir := range []Direction{TopDown, BottomUp, LeftRight, RightLeft} {
			g, index := testGraph(t, tt.edges...)
			g.Dir = dir
			g.Groups = tt.groups(index)
			for _, e := range tt.groupEdges {
				e.Length = 1
				g.Edges = append(g.Edges, e)
			}
			if err := check(g); err != nil {
				t.Fatal(err)
			}
			l, err := layOut(g)
			if err != nil {
				t.Fatalf("%s: layOut() error = %v", tt.name, err)
			}
			checkFrames(t, tt.name, l)
		}
	}
}

// checkFrames checks that each frame takes in the vertices in it and the
// frames nested in it, and keeps out the rest.
func checkFrames(t *testing.T, name string, l *layout) {
	t.Helper()
	type rect struct{ x0, y0, x1, y1 int }
	frameRect := func(f frame) rect { return rect{f.x, f.y, f.x + f.w - 1, f.y + f.h - 1} }
	vertRect := func(v *vertex) rect { return rect{v.cx - l.left(v), v.y, v.cx + l.right(v), v.y + v.h - 1} }
	inside := func(a, b rect) bool { return a.x0 > b.x0 && a.x1 < b.x1 && a.y0 > b.y0 && a.y1 < b.y1 }
	apart := func(a, b rect) bool { return a.x1 < b.x0 || b.x1 < a.x0 || a.y1 < b.y0 || b.y1 < a.y0 }

	for f, fr := range l.frames {
		fRect := frameRect(fr)
		if fr.x < 0 || fr.y < 0 || fr.x+fr.w > l.w || fr.y+fr.h > l.h {
			t.Errorf("%s: frame %d at %+v is outside the layout's %dx%d", name, f, fRect, l.w, l.h)
		}
		if l.orient.sideways && fr.h < fr.titleSpan() || !l.orient.sideways && fr.w < fr.titleSpan() {
			t.Errorf("%s: frame %d at %+v is too narrow for its title %q", name, f, fRect, fr.title)
		}
		for v := range l.verts {
			vert := &l.verts[v]
			if vert.isFiller() {
				continue
			}
			in := l.lca(vert.frame, f) == f
			if vRect := vertRect(vert); in && !inside(vRect, fRect) {
				t.Errorf("%s: vertex %d at %+v isn't inside its frame %d at %+v", name, v, vRect, f, fRect)
			} else if !in && !apart(vRect, fRect) {
				t.Errorf("%s: vertex %d at %+v overlaps frame %d at %+v", name, v, vRect, f, fRect)
			}
		}
		for c, ch := range l.frames {
			switch cRect := frameRect(ch); {
			case c == f:
			case l.lca(c, f) == f:
				if !inside(cRect, fRect) {
					t.Errorf("%s: frame %d at %+v isn't inside frame %d at %+v", name, c, cRect, f, fRect)
				}
			case l.lca(c, f) != c && !apart(cRect, fRect):
				t.Errorf("%s: frames %d at %+v and %d at %+v overlap", name, c, cRect, f, fRect)
			}
		}
	}
}
