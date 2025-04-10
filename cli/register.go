package cli

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

func InitCommands(ctx context.Context) int {

	subcommands.Register(&configCmd{}, "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&helpCmd{}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
