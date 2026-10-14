package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/internal/config"

	"github.com/google/subcommands"
)

type versionCmd struct{}

var _ subcommands.Command = (*versionCmd)(nil)

func (*versionCmd) Name() string { return "version" }

func (*versionCmd) Synopsis() string { return "nbx version" }

func (*versionCmd) Usage() string {
	return `nbx version:
show nbx version.
`
}

func (*versionCmd) SetFlags(f *flag.FlagSet) {}

func (c *versionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {

	fmt.Println("nbx", config.CurrentVersion)

	return subcommands.ExitSuccess
}
