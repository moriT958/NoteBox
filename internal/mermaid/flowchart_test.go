package mermaid

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseFlowchart(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want flowchart
	}{
		{
			name: "declarations",
			src:  "graph",
			want: flowchart{
				dir:   flowTopDown,
				index: map[string]int{},
			},
		},
		{
			name: "directions",
			src:  "flowchart LR\nA",
			want: flowchart{
				dir: flowLeftRight,
				nodes: []node{
					{id: "A", label: "A"},
				},
				index: map[string]int{"A": 0},
			},
		},
		{
			name: "statements on the declaration line",
			src:  "graph BT;A-->B;",
			want: flowchart{
				dir: flowBottomUp,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "B"},
				},
				index: map[string]int{"A": 0, "B": 1},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "comments and front matter",
			src:  "---\ntitle: t\n---\n%%{init: {}}%%\ngraph RL\n%% A-->C\nA-->B %% trailing",
			want: flowchart{
				dir: flowRightLeft,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "B"},
				},
				index: map[string]int{"A": 0, "B": 1},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "shapes",
			src: strings.Join([]string{
				"graph TD",
				"a[rect] --- b(round) --- c([stadium]) --- d[[subroutine]] --- e[(cylinder)]",
				"f((circle)) --- g(((double))) --- h>odd] --- i{diamond} --- j{{hexagon}}",
				`k[/lean/] --- l[\lean\] --- m[/trap\] --- n[\trap/] --- o(-ellipse-)`,
			}, "\n"),
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "a", label: "rect"},
					{id: "b", label: "round", shape: shapeRound},
					{id: "c", label: "stadium", shape: shapeRound},
					{id: "d", label: "subroutine"},
					{id: "e", label: "cylinder", shape: shapeRound},
					{id: "f", label: "circle", shape: shapeRound},
					{id: "g", label: "double", shape: shapeRound},
					{id: "h", label: "odd"},
					{id: "i", label: "diamond", shape: shapeDiamond},
					{id: "j", label: "hexagon", shape: shapeDiamond},
					{id: "k", label: "lean"},
					{id: "l", label: "lean"},
					{id: "m", label: "trap"},
					{id: "n", label: "trap"},
					{id: "o", label: "ellipse", shape: shapeRound},
				},
				index: map[string]int{"a": 0, "b": 1, "c": 2, "d": 3, "e": 4, "f": 5, "g": 6, "h": 7, "i": 8, "j": 9, "k": 10, "l": 11, "m": 12, "n": 13, "o": 14},
				edges: []edge{
					{from: 0, to: 1, length: 1},
					{from: 1, to: 2, length: 1},
					{from: 2, to: 3, length: 1},
					{from: 3, to: 4, length: 1},
					{from: 5, to: 6, length: 1},
					{from: 6, to: 7, length: 1},
					{from: 7, to: 8, length: 1},
					{from: 8, to: 9, length: 1},
					{from: 10, to: 11, length: 1},
					{from: 11, to: 12, length: 1},
					{from: 12, to: 13, length: 1},
					{from: 13, to: 14, length: 1},
				},
			},
		},
		{
			name: "labels",
			src:  "graph TD\nA[\"a [quoted] label\"] --> B[\"`**bold**`\"]\nC[ノート<br>作成]",
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "a [quoted] label"},
					{id: "B", label: "bold"},
					{id: "C", label: "ノート\n作成"},
				},
				index: map[string]int{"A": 0, "B": 1, "C": 2},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "a later label replaces the ID",
			src:  "graph TD\nA --> B\nB(Done)\nB --> C",
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "Done", shape: shapeRound},
					{id: "C", label: "C"},
				},
				index: map[string]int{"A": 0, "B": 1, "C": 2},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
					{from: 1, to: 2, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "IDs",
			src:  "graph TD\nnode-1 --> a.b --> foo/bar --> 日本語 --> x_y",
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "node-1", label: "node-1"},
					{id: "a.b", label: "a.b"},
					{id: "foo/bar", label: "foo/bar"},
					{id: "日本語", label: "日本語"},
					{id: "x_y", label: "x_y"},
				},
				index: map[string]int{"node-1": 0, "a.b": 1, "foo/bar": 2, "日本語": 3, "x_y": 4},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
					{from: 1, to: 2, toHead: headArrow, length: 1},
					{from: 2, to: 3, toHead: headArrow, length: 1},
					{from: 3, to: 4, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "classes",
			src:  "graph TD\nA:::warn --> B[b]:::ok-2\nclassDef warn fill:#f00\nclass B ok-2\nstyle A fill:#f9f,stroke:#333\nlinkStyle 0 stroke:red\nclick A callback\ndirection LR",
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "b"},
				},
				index: map[string]int{"A": 0, "B": 1},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "accessibility",
			src:  "graph TD\naccTitle: a; title\naccDescr: a description\naccDescr {\n  many\n  lines\n}\nA",
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "A"},
				},
				index: map[string]int{"A": 0},
			},
		},
		{
			name: "chains and groups",
			src:  "graph TD\nA --> B --> C\nD & E --> F & G",
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "B"},
					{id: "C", label: "C"},
					{id: "D", label: "D"},
					{id: "E", label: "E"},
					{id: "F", label: "F"},
					{id: "G", label: "G"},
				},
				index: map[string]int{"A": 0, "B": 1, "C": 2, "D": 3, "E": 4, "F": 5, "G": 6},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
					{from: 1, to: 2, toHead: headArrow, length: 1},
					{from: 3, to: 5, toHead: headArrow, length: 1},
					{from: 3, to: 6, toHead: headArrow, length: 1},
					{from: 4, to: 5, toHead: headArrow, length: 1},
					{from: 4, to: 6, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "links",
			src: strings.Join([]string{
				"graph TD",
				"A --> B",
				"A --- B",
				"A --o B",
				"A --x B",
				"A <--> B",
				"A o--o B",
				"A x--x B",
				"A ----> B",
				"A ---- B",
				"A -.-> B",
				"A -.- B",
				"A -...-> B",
				"A <-.-> B",
				"A ==> B",
				"A === B",
				"A ====> B",
				"A <==> B",
				"A ~~~ B",
				"A ~~~~ B",
				"A-->B",
			}, "\n"),
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "B"},
				},
				index: map[string]int{"A": 0, "B": 1},
				edges: []edge{
					{from: 0, to: 1, toHead: headArrow, length: 1},
					{from: 0, to: 1, length: 1},
					{from: 0, to: 1, toHead: headCircle, length: 1},
					{from: 0, to: 1, toHead: headCross, length: 1},
					{from: 0, to: 1, fromHead: headArrow, toHead: headArrow, length: 1},
					{from: 0, to: 1, fromHead: headCircle, toHead: headCircle, length: 1},
					{from: 0, to: 1, fromHead: headCross, toHead: headCross, length: 1},
					{from: 0, to: 1, toHead: headArrow, length: 3},
					{from: 0, to: 1, length: 2},
					{from: 0, to: 1, toHead: headArrow, style: lineDotted, length: 1},
					{from: 0, to: 1, style: lineDotted, length: 1},
					{from: 0, to: 1, toHead: headArrow, style: lineDotted, length: 3},
					{from: 0, to: 1, fromHead: headArrow, toHead: headArrow, style: lineDotted, length: 1},
					{from: 0, to: 1, toHead: headArrow, style: lineThick, length: 1},
					{from: 0, to: 1, style: lineThick, length: 1},
					{from: 0, to: 1, toHead: headArrow, style: lineThick, length: 3},
					{from: 0, to: 1, fromHead: headArrow, toHead: headArrow, style: lineThick, length: 1},
					{from: 0, to: 1, invisible: true, length: 1},
					{from: 0, to: 1, invisible: true, length: 2},
					{from: 0, to: 1, toHead: headArrow, length: 1},
				},
			},
		},
		{
			name: "link labels",
			src: strings.Join([]string{
				"graph TD",
				"A -->|yes| B",
				"A ---|\"a | b\"| B",
				"A -.->|maybe| B",
				"A == heavy ==> B",
				"A -- text --> B",
				"A -- open --- B",
				"A -- long ----> B",
				"A -. dotted .-> B",
				"A -. dotted -.-> B",
				"A -- a-b --x B",
				"A -- 日本語 --> B",
				"A -->|user@example.com| B",
			}, "\n"),
			want: flowchart{
				dir: flowTopDown,
				nodes: []node{
					{id: "A", label: "A"},
					{id: "B", label: "B"},
				},
				index: map[string]int{"A": 0, "B": 1},
				edges: []edge{
					{from: 0, to: 1, label: "yes", toHead: headArrow, length: 1},
					{from: 0, to: 1, label: "a | b", length: 1},
					{from: 0, to: 1, label: "maybe", toHead: headArrow, style: lineDotted, length: 1},
					{from: 0, to: 1, label: "heavy", toHead: headArrow, style: lineThick, length: 1},
					{from: 0, to: 1, label: "text", toHead: headArrow, length: 1},
					{from: 0, to: 1, label: "open", length: 1},
					{from: 0, to: 1, label: "long", toHead: headArrow, length: 3},
					{from: 0, to: 1, label: "dotted", toHead: headArrow, style: lineDotted, length: 1},
					{from: 0, to: 1, label: "dotted", toHead: headArrow, style: lineDotted, length: 1},
					{from: 0, to: 1, label: "a-b", toHead: headCross, length: 1},
					{from: 0, to: 1, label: "日本語", toHead: headArrow, length: 1},
					{from: 0, to: 1, label: "user@example.com", toHead: headArrow, length: 1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc, err := parseFlowchart(tt.src)
			if err != nil {
				t.Fatalf("parseFlowchart() error = %v", err)
			}
			if !reflect.DeepEqual(*fc, tt.want) {
				t.Errorf("got\n%+v\nwant\n%+v", *fc, tt.want)
			}
		})
	}
}

func TestParseFlowchartErrors(t *testing.T) {
	tests := []struct {
		src  string
		want error
	}{
		{"graph TD\nsubgraph one\nA\nend", ErrUnsupported},
		{"graph TD\nA@{ shape: rect }", ErrUnsupported},
		{"graph TD\nA e1@--> B", ErrUnsupported},
		{"graph XY", ErrSyntax},
		{"graph TD extra", ErrSyntax},
		{"graph TD\nA[unclosed", ErrSyntax},
		{"graph TD\nA -->", ErrSyntax},
		{"graph TD\nA -- unclosed label", ErrSyntax},
		{"graph TD\nA -> B", ErrSyntax},
		{"graph TD\nA ~~ B", ErrSyntax},
		{"graph TD\nA B", ErrSyntax},
		{"graph TD\n--> B", ErrSyntax},
		{"graph TD\nA -->|unclosed B", ErrSyntax},
	}
	for _, tt := range tests {
		if _, err := parseFlowchart(tt.src); !errors.Is(err, tt.want) {
			t.Errorf("parseFlowchart(%q) error = %v, want %v", tt.src, err, tt.want)
		}
	}
}

func TestDrawFlowchartErrors(t *testing.T) {
	tests := []struct {
		src  string
		want error
	}{
		{"graph TD", ErrSyntax},
		{"graph TD\nclassDef a fill:#f00", ErrSyntax},
	}
	for _, tt := range tests {
		if _, err := drawFlowchart(tt.src); !errors.Is(err, tt.want) {
			t.Errorf("drawFlowchart(%q) error = %v, want %v", tt.src, err, tt.want)
		}
	}
}
