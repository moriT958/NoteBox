package cli

import (
	"context"
	"flag"

	"github.com/google/subcommands"

	"notebox/internal/note"
)

func InitCommands(ctx context.Context, repo note.BoxRepository) int {

	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&listCmd{repo: repo}, "")
	subcommands.Register(&pruneCmd{repo: repo}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
