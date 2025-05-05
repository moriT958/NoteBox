package cli

import (
	"context"
	"flag"
	"fmt"

	"NoteBox.tmp/config"
	"github.com/google/subcommands"
)

type versionCmd struct{}

var _ subcommands.Command = (*versionCmd)(nil)

func (*versionCmd) Name() string { return "version" }

func (*versionCmd) Synopsis() string { return "notebox version" }

func (*versionCmd) Usage() string {
	return `note version:
show notebox version.
`
}

func (*versionCmd) SetFlags(f *flag.FlagSet) {}

func (c *versionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {

	fmt.Println("notebox", config.CurrentVersion)

	return subcommands.ExitSuccess
}
