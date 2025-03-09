package cli

import (
	"context"
	"flag"
	"log"
	"notebox/config"
	"notebox/store"

	"github.com/google/subcommands"
)

func InitCommands(ctx context.Context, cfg *config.Config) int {
	store, err := store.NewNoteStore(cfg.MetaDataPath)
	if err != nil {
		log.Println(err)
		return int(subcommands.ExitFailure)
	}

	subcommands.Register(&newCmd{cfg: cfg, store: store}, "")
	subcommands.Register(&lsCmd{cfg: cfg, store: store}, "")
	subcommands.Register(&editCmd{cfg: cfg, store: store}, "")
	subcommands.Register(&rmCmd{cfg: cfg, store: store}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
