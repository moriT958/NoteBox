package styles

import (
	"fmt"
	"regexp"
)

// Palette is the colors of the TUI, by what they're for. A theme is a palette.
type Palette struct {
	// Primary frames panels that don't have the focus, and underlines the
	// header.
	Primary string `json:"primary"`
	// Accent frames the focused panel and marks titles.
	Accent string `json:"accent"`
	// Cursor marks the selected row, CursorMuted its path.
	Cursor      string `json:"cursor"`
	CursorMuted string `json:"cursor_muted"`
	Muted       string `json:"muted"`
	TabBarBg    string `json:"tab_bar_bg"`
	// ButtonFg is the text of the dialog buttons.
	ButtonFg string `json:"button_fg"`
	Success  string `json:"success"`
	// Danger colors the cancel button and errors.
	Danger string `json:"danger"`
	// SelectionFg and SelectionBg mark the text selected with the mouse in
	// the preview.
	SelectionFg string `json:"selection_fg"`
	SelectionBg string `json:"selection_bg"`

	MermaidBorder    string `json:"mermaid_border"`
	MermaidText      string `json:"mermaid_text"`
	MermaidEdge      string `json:"mermaid_edge"`
	MermaidEdgeLabel string `json:"mermaid_edge_label"`

	// Markdown is the base style of the rendered markdown: "dark" or "light".
	Markdown string `json:"markdown"`
}

// token is one value of a palette, with its name in theme files.
type token struct {
	name  string
	value *string
	check func(string) error
}

// tokens is the one list of the palette's values; loading a theme and
// checking its values go by it.
func (p *Palette) tokens() []token {
	return []token{
		{"primary", &p.Primary, checkColor},
		{"accent", &p.Accent, checkColor},
		{"cursor", &p.Cursor, checkColor},
		{"cursor_muted", &p.CursorMuted, checkColor},
		{"muted", &p.Muted, checkColor},
		{"tab_bar_bg", &p.TabBarBg, checkColor},
		{"button_fg", &p.ButtonFg, checkColor},
		{"success", &p.Success, checkColor},
		{"danger", &p.Danger, checkColor},
		{"selection_fg", &p.SelectionFg, checkColor},
		{"selection_bg", &p.SelectionBg, checkColor},
		{"mermaid_border", &p.MermaidBorder, checkColor},
		{"mermaid_text", &p.MermaidText, checkColor},
		{"mermaid_edge", &p.MermaidEdge, checkColor},
		{"mermaid_edge_label", &p.MermaidEdgeLabel, checkColor},
		{"markdown", &p.Markdown, checkMarkdown},
	}
}

var hexColor = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func checkColor(s string) error {
	if !hexColor.MatchString(s) {
		return fmt.Errorf("invalid color %q, use #rgb or #rrggbb", s)
	}
	return nil
}

func checkMarkdown(s string) error {
	if s != "dark" && s != "light" {
		return fmt.Errorf(`invalid markdown style %q, use "dark" or "light"`, s)
	}
	return nil
}

// DarkPalette returns the built-in dark theme.
func DarkPalette() Palette {
	return Palette{
		Primary:          "#737994",
		Accent:           "#babbf1",
		Cursor:           "#f2d5cf",
		CursorMuted:      "#c7a89e",
		Muted:            "#51576d",
		TabBarBg:         "#232634",
		ButtonFg:         "#414559",
		Success:          "#99d1db",
		Danger:           "#ea999c",
		SelectionFg:      "#303446",
		SelectionBg:      "#8caaee",
		MermaidBorder:    "#8caaee",
		MermaidText:      "#c6d0f5",
		MermaidEdge:      "#949cbb",
		MermaidEdgeLabel: "#a5adce",
		Markdown:         "dark",
	}
}

// LightPalette returns the built-in light theme.
func LightPalette() Palette {
	return Palette{
		Primary:          "#b89988",
		Accent:           "#fe640b",
		Cursor:           "#dc8a78",
		CursorMuted:      "#b58c7e",
		Muted:            "#8c8fa1",
		TabBarBg:         "#dce0e8",
		ButtonFg:         "#eff1f5",
		Success:          "#40a02b",
		Danger:           "#d20f39",
		SelectionFg:      "#4c4f69",
		SelectionBg:      "#bcd0f7",
		MermaidBorder:    "#1e66f5",
		MermaidText:      "#4c4f69",
		MermaidEdge:      "#7c7f93",
		MermaidEdgeLabel: "#6c6f85",
		Markdown:         "light",
	}
}

// builtinNames are the names of the built-in themes.
var builtinNames = []string{"dark", "light"}

// builtin returns the built-in theme of that name.
func builtin(name string) (Palette, bool) {
	switch name {
	case "dark":
		return DarkPalette(), true
	case "light":
		return LightPalette(), true
	}
	return Palette{}, false
}
