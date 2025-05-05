package cli

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

func InitCommands(ctx context.Context) int {

	subcommands.Register(&versionCmd{}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
