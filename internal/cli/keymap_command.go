package cli

import (
	"context"
	"flag"
	"slices"
	"strings"

	"notebox/internal/tui/keymap"

	"github.com/google/subcommands"
)

type keymapListCmd struct{}

var _ subcommands.Command = (*keymapListCmd)(nil)

func (*keymapListCmd) Name() string { return "list" }

func (*keymapListCmd) Synopsis() string { return "list the default keys of every action" }

func (*keymapListCmd) Usage() string {
	return `nbx keymap list:
show the default keys of every action in the TUI. on a terminal it is a table,
otherwise it is one action per line as "<action>	<keys>", the keys separated by
spaces. <action> is what the "keymap" of config.json takes.
`
}

func (*keymapListCmd) SetFlags(f *flag.FlagSet) {}

func (c *keymapListCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}

	defaults := keymap.Default()
	keys := defaults.Keys()
	ids := make([]string, 0, len(keys))
	for id := range keys {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	rows := make([][]string, len(ids))
	for i, id := range ids {
		rows[i] = []string{id, strings.Join(keys[id], " ")}
	}
	printTable([]string{"Action", "Keys"}, rows, nil)

	return subcommands.ExitSuccess
}
