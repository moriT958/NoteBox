package graph

import (
	"errors"
	"testing"

	"notebox/internal/mermaid/internal/diagram"
)

func TestDrawInvalidGroups(t *testing.T) {
	for _, groups := range [][]Group{
		{{Nodes: []int{2}}},
		{{Nodes: []int{0}}, {Nodes: []int{0}}},
		{{Groups: []int{1}}, {Groups: []int{0}}},
		{{Groups: []int{0}}},
		{{Groups: []int{1}}, {}, {Groups: []int{1}}},
	} {
		g := &Graph{Nodes: []Node{{Label: "A"}, {Label: "B"}}, Groups: groups}
		_, err := Draw(g)
		if err == nil || errors.Is(err, diagram.ErrSyntax) || errors.Is(err, diagram.ErrTooLarge) || errors.Is(err, diagram.ErrUnsupported) {
			t.Errorf("Draw() with groups %+v error = %v, want an error of its own", groups, err)
		}
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
			if err := checkGroups(g); err != nil {
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
