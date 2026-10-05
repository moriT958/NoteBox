package flowchart

import (
	"errors"
	"notebox/internal/mermaid/diagram"
	"notebox/internal/mermaid/graph"
	"reflect"
	"strings"
	"testing"
)

func TestParseFlowchart(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want graph.Graph
	}{
		{
			name: "declarations",
			src:  "graph",
			want: graph.Graph{
				Dir: graph.TopDown,
			},
		},
		{
			name: "directions",
			src:  "flowchart LR\nA",
			want: graph.Graph{
				Dir: graph.LeftRight,
				Nodes: []graph.Node{
					{Label: "A"},
				},
			},
		},
		{
			name: "statements on the declaration line",
			src:  "graph BT;A-->B;",
			want: graph.Graph{
				Dir: graph.BottomUp,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "B"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
				},
			},
		},
		{
			name: "comments and front matter",
			src:  "---\ntitle: t\n---\n%%{init: {}}%%\ngraph RL\n%% A-->C\nA-->B %% trailing",
			want: graph.Graph{
				Dir: graph.RightLeft,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "B"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
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
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "rect"},
					{Label: "round", Shape: graph.ShapeRound},
					{Label: "stadium", Shape: graph.ShapeRound},
					{Label: "subroutine"},
					{Label: "cylinder", Shape: graph.ShapeRound},
					{Label: "circle", Shape: graph.ShapeRound},
					{Label: "double", Shape: graph.ShapeRound},
					{Label: "odd"},
					{Label: "diamond", Shape: graph.ShapeDiamond},
					{Label: "hexagon", Shape: graph.ShapeDiamond},
					{Label: "lean"},
					{Label: "lean"},
					{Label: "trap"},
					{Label: "trap"},
					{Label: "ellipse", Shape: graph.ShapeRound},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, Length: 1},
					{From: 1, To: 2, Length: 1},
					{From: 2, To: 3, Length: 1},
					{From: 3, To: 4, Length: 1},
					{From: 5, To: 6, Length: 1},
					{From: 6, To: 7, Length: 1},
					{From: 7, To: 8, Length: 1},
					{From: 8, To: 9, Length: 1},
					{From: 10, To: 11, Length: 1},
					{From: 11, To: 12, Length: 1},
					{From: 12, To: 13, Length: 1},
					{From: 13, To: 14, Length: 1},
				},
			},
		},
		{
			name: "labels",
			src:  "graph TD\nA[\"a [quoted] label\"] --> B[\"`**bold**`\"]\nC[ノート<br>作成]",
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "a [quoted] label"},
					{Label: "bold"},
					{Label: "ノート\n作成"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
				},
			},
		},
		{
			name: "a later label replaces the ID",
			src:  "graph TD\nA --> B\nB(Done)\nB --> C",
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "Done", Shape: graph.ShapeRound},
					{Label: "C"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
					{From: 1, To: 2, ToHead: graph.HeadArrow, Length: 1},
				},
			},
		},
		{
			name: "IDs",
			src:  "graph TD\nnode-1 --> a.b --> foo/bar --> 日本語 --> x_y",
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "node-1"},
					{Label: "a.b"},
					{Label: "foo/bar"},
					{Label: "日本語"},
					{Label: "x_y"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
					{From: 1, To: 2, ToHead: graph.HeadArrow, Length: 1},
					{From: 2, To: 3, ToHead: graph.HeadArrow, Length: 1},
					{From: 3, To: 4, ToHead: graph.HeadArrow, Length: 1},
				},
			},
		},
		{
			name: "classes",
			src:  "graph TD\nA:::warn --> B[b]:::ok-2\nclassDef warn fill:#f00\nclass B ok-2\nstyle A fill:#f9f,stroke:#333\nlinkStyle 0 stroke:red\nclick A callback\ndirection LR",
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "b"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
				},
			},
		},
		{
			name: "accessibility",
			src:  "graph TD\naccTitle: a; title\naccDescr: a description\naccDescr {\n  many\n  lines\n}\nA",
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "A"},
				},
			},
		},
		{
			name: "chains and groups",
			src:  "graph TD\nA --> B --> C\nD & E --> F & G",
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "B"},
					{Label: "C"},
					{Label: "D"},
					{Label: "E"},
					{Label: "F"},
					{Label: "G"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
					{From: 1, To: 2, ToHead: graph.HeadArrow, Length: 1},
					{From: 3, To: 5, ToHead: graph.HeadArrow, Length: 1},
					{From: 3, To: 6, ToHead: graph.HeadArrow, Length: 1},
					{From: 4, To: 5, ToHead: graph.HeadArrow, Length: 1},
					{From: 4, To: 6, ToHead: graph.HeadArrow, Length: 1},
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
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "B"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
					{From: 0, To: 1, Length: 1},
					{From: 0, To: 1, ToHead: graph.HeadCircle, Length: 1},
					{From: 0, To: 1, ToHead: graph.HeadCross, Length: 1},
					{From: 0, To: 1, FromHead: graph.HeadArrow, ToHead: graph.HeadArrow, Length: 1},
					{From: 0, To: 1, FromHead: graph.HeadCircle, ToHead: graph.HeadCircle, Length: 1},
					{From: 0, To: 1, FromHead: graph.HeadCross, ToHead: graph.HeadCross, Length: 1},
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 3},
					{From: 0, To: 1, Length: 2},
					{From: 0, To: 1, ToHead: graph.HeadArrow, Style: diagram.LineDotted, Length: 1},
					{From: 0, To: 1, Style: diagram.LineDotted, Length: 1},
					{From: 0, To: 1, ToHead: graph.HeadArrow, Style: diagram.LineDotted, Length: 3},
					{From: 0, To: 1, FromHead: graph.HeadArrow, ToHead: graph.HeadArrow, Style: diagram.LineDotted, Length: 1},
					{From: 0, To: 1, ToHead: graph.HeadArrow, Style: diagram.LineThick, Length: 1},
					{From: 0, To: 1, Style: diagram.LineThick, Length: 1},
					{From: 0, To: 1, ToHead: graph.HeadArrow, Style: diagram.LineThick, Length: 3},
					{From: 0, To: 1, FromHead: graph.HeadArrow, ToHead: graph.HeadArrow, Style: diagram.LineThick, Length: 1},
					{From: 0, To: 1, Invisible: true, Length: 1},
					{From: 0, To: 1, Invisible: true, Length: 2},
					{From: 0, To: 1, ToHead: graph.HeadArrow, Length: 1},
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
			want: graph.Graph{
				Dir: graph.TopDown,
				Nodes: []graph.Node{
					{Label: "A"},
					{Label: "B"},
				},
				Edges: []graph.Edge{
					{From: 0, To: 1, Label: "yes", ToHead: graph.HeadArrow, Length: 1},
					{From: 0, To: 1, Label: "a | b", Length: 1},
					{From: 0, To: 1, Label: "maybe", ToHead: graph.HeadArrow, Style: diagram.LineDotted, Length: 1},
					{From: 0, To: 1, Label: "heavy", ToHead: graph.HeadArrow, Style: diagram.LineThick, Length: 1},
					{From: 0, To: 1, Label: "text", ToHead: graph.HeadArrow, Length: 1},
					{From: 0, To: 1, Label: "open", Length: 1},
					{From: 0, To: 1, Label: "long", ToHead: graph.HeadArrow, Length: 3},
					{From: 0, To: 1, Label: "dotted", ToHead: graph.HeadArrow, Style: diagram.LineDotted, Length: 1},
					{From: 0, To: 1, Label: "dotted", ToHead: graph.HeadArrow, Style: diagram.LineDotted, Length: 1},
					{From: 0, To: 1, Label: "a-b", ToHead: graph.HeadCross, Length: 1},
					{From: 0, To: 1, Label: "日本語", ToHead: graph.HeadArrow, Length: 1},
					{From: 0, To: 1, Label: "user@example.com", ToHead: graph.HeadArrow, Length: 1},
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
		{"graph TD\nsubgraph one\nA\nend", diagram.ErrUnsupported},
		{"graph TD\nA@{ Shape: rect }", diagram.ErrUnsupported},
		{"graph TD\nA e1@--> B", diagram.ErrUnsupported},
		{"graph XY", diagram.ErrSyntax},
		{"graph TD extra", diagram.ErrSyntax},
		{"graph TD\nA[unclosed", diagram.ErrSyntax},
		{"graph TD\nA -->", diagram.ErrSyntax},
		{"graph TD\nA -- unclosed label", diagram.ErrSyntax},
		{"graph TD\nA -> B", diagram.ErrSyntax},
		{"graph TD\nA ~~ B", diagram.ErrSyntax},
		{"graph TD\nA B", diagram.ErrSyntax},
		{"graph TD\n--> B", diagram.ErrSyntax},
		{"graph TD\nA -->|unclosed B", diagram.ErrSyntax},
	}
	for _, tt := range tests {
		if _, err := parseFlowchart(tt.src); !errors.Is(err, tt.want) {
			t.Errorf("parseFlowchart(%q) error = %v, want %v", tt.src, err, tt.want)
		}
	}
}

func TestDrawErrors(t *testing.T) {
	tests := []struct {
		src  string
		want error
	}{
		{"graph TD", diagram.ErrSyntax},
		{"graph TD\nclassDef a fill:#f00", diagram.ErrSyntax},
	}
	for _, tt := range tests {
		if _, err := Draw(tt.src); !errors.Is(err, tt.want) {
			t.Errorf("Draw(%q) error = %v, want %v", tt.src, err, tt.want)
		}
	}
}

func TestParseFlowchartLines(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"graph XY", "line 1: "},
		{"\n\ngraph TD", ""},
		{"---\ntitle: t\n---\ngraph TD\nA --> B\n\n%% comment\nA -> B", "line 8: "},
		{"graph TD\nA --> B; C -> D", "line 2: "},
		{"graph TD\naccDescr {\n  many\n  lines\n}\nA[unclosed", "line 6: "},
	}
	for _, tt := range tests {
		_, err := parseFlowchart(tt.src)
		switch {
		case tt.want == "" && err != nil:
			t.Errorf("parseFlowchart(%q) error = %v, want none", tt.src, err)
		case tt.want != "" && (err == nil || !strings.HasPrefix(err.Error(), tt.want)):
			t.Errorf("parseFlowchart(%q) error = %v, want it to start with %q", tt.src, err, tt.want)
		}
	}
}
