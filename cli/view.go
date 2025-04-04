package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/subcommands"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type viewCmd struct{}

var _ subcommands.Command = (*viewCmd)(nil)

func (*viewCmd) Name() string { return "view" }

func (*viewCmd) Synopsis() string { return "preview note contents" }

func (*viewCmd) Usage() string {
	return `note view <id>:
render markdown file and preview on terminal.`
}

func (*viewCmd) SetFlags(f *flag.FlagSet) {}

func (c *viewCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	id, err := getIdArg(f.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	note, err := Nr.FindByID(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	content, err := os.ReadFile(note.GetFilePath())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	// Bubble teaで表示
	model, err := newPreviewModel(string(content))
	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		return subcommands.ExitFailure
	}

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Bummer, there's been an error:", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

type previewModel struct {
	viewport viewport.Model
}

func newPreviewModel(content string) (*previewModel, error) {
	const width = 150

	vp := viewport.New(width, 30)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//
	const glamourGutter = 2
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}

	str, err := renderer.Render(content)
	if err != nil {
		return nil, err
	}

	vp.SetContent(str)

	return &previewModel{
		viewport: vp,
	}, nil
}

func (e previewModel) Init() tea.Cmd {
	return nil
}

func (e previewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return e, tea.Quit
		default:
			var cmd tea.Cmd
			e.viewport, cmd = e.viewport.Update(msg)
			return e, cmd
		}
	default:
		return e, nil
	}
}

func (e previewModel) View() string {
	return e.viewport.View() + e.helpView()
}

func (e previewModel) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
