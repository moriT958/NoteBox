package cli

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

type helpCmd struct{}

var _ subcommands.Command = (*helpCmd)(nil)

func (*helpCmd) Name() string { return "help" }

func (*helpCmd) Synopsis() string { return "show help" }

func (*helpCmd) Usage() string {
	return `help:
Shows a list of commands
`
}

func (*helpCmd) SetFlags(f *flag.FlagSet) {}

func (c *helpCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	return subcommands.HelpCommand().Execute(ctx, f, args...)
}
