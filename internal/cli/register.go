package cli

import (
	"context"
	"flag"
	"notebox/internal/core/box"

	"github.com/google/subcommands"
)

func InitCommands(ctx context.Context, boxes *box.BoxService) int {

	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&listCmd{boxes: boxes}, "")
	subcommands.Register(&pruneCmd{boxes: boxes}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
