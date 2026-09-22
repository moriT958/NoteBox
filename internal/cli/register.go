package cli

import (
	"context"
	"flag"
	"notebox/internal/notedeprecated"

	"github.com/google/subcommands"
)

func InitCommands(ctx context.Context, repo notedeprecated.BoxRepository) int {

	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&listCmd{repo: repo}, "")
	subcommands.Register(&pruneCmd{repo: repo}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
