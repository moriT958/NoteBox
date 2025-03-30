package cli

import (
	"context"
	"flag"
	"notebox/models"

	"github.com/google/subcommands"
)

var (
	Nr *models.NoteRepository
)

func InitCommands(ctx context.Context) int {

	subcommands.Register(&newCmd{}, "")
	subcommands.Register(&lsCmd{}, "")
	subcommands.Register(&editCmd{}, "")
	subcommands.Register(&rmCmd{}, "")
	subcommands.Register(&configCmd{}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
