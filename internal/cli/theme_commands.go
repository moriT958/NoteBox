package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"notebox/internal/config"
	"notebox/internal/tui/styles"

	"github.com/google/subcommands"
)

type themeListCmd struct{}

var _ subcommands.Command = (*themeListCmd)(nil)

func (*themeListCmd) Name() string { return "list" }

func (*themeListCmd) Synopsis() string { return "list the themes that can be used" }

func (*themeListCmd) Usage() string {
	return `nbx theme list:
show the themes that can be set as "theme" in config.json: the built-in themes
and the files in ~/.notebox/themes. on a terminal it is a table, otherwise it
is one theme per line as "<name>	<builtin|file>". a file takes precedence over
a built-in theme of the same name.
`
}

func (*themeListCmd) SetFlags(f *flag.FlagSet) {}

func (c *themeListCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}

	dir, err := config.ThemesDir()
	if err != nil {
		return fail(fmt.Errorf("failed to find the themes dir: %w", err))
	}
	themes, err := styles.ListThemes(dir)
	if err != nil {
		return fail(err)
	}

	rows := make([][]string, len(themes))
	for i, t := range themes {
		source := "builtin"
		if t.File {
			source = "file"
		}
		rows[i] = []string{t.Name, source}
	}
	printTable([]string{"Theme", "Source"}, rows, nil)

	return subcommands.ExitSuccess
}

type themeExportCmd struct{}

var _ subcommands.Command = (*themeExportCmd)(nil)

func (*themeExportCmd) Name() string { return "export" }

func (*themeExportCmd) Synopsis() string { return "print a built-in theme as a theme file" }

func (*themeExportCmd) Usage() string {
	return `nbx theme export <name>:
print the theme file of a built-in theme (dark or light) to start a theme of
your own from, for example:
  mkdir -p ~/.notebox/themes
  nbx theme export dark > ~/.notebox/themes/mine.json
then set "theme": "mine" in config.json.
`
}

func (*themeExportCmd) SetFlags(f *flag.FlagSet) {}

func (c *themeExportCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 1)
	if !ok {
		return subcommands.ExitUsageError
	}

	data, err := styles.ExportTheme(pos[0])
	if err != nil {
		return fail(err)
	}
	os.Stdout.Write(data)

	return subcommands.ExitSuccess
}
