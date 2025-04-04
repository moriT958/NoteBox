package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

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

	fmt.Printf("Received ID: %d\n", id)

	return subcommands.ExitSuccess
}
