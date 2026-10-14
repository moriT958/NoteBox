package cli

import (
	"context"
	"flag"
	"notebox/internal/core/box"

	"github.com/google/subcommands"
)

func InitCommands(ctx context.Context, boxes *box.BoxService) int {

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&group{
		name:     "box",
		synopsis: "manage boxes",
		cmds: []subcommands.Command{
			&boxListCmd{boxes: boxes},
			&boxAddCmd{boxes: boxes},
			&boxRenameCmd{boxes: boxes},
			&boxRelinkCmd{boxes: boxes},
			&boxRmCmd{boxes: boxes},
			&boxPruneCmd{boxes: boxes},
		},
	}, "")
	subcommands.Register(&group{
		name:     "note",
		synopsis: "manage notes in a box",
		cmds: []subcommands.Command{
			&noteListCmd{boxes: boxes},
			&noteNewCmd{boxes: boxes},
			&noteRenameCmd{boxes: boxes},
			&noteViewCmd{boxes: boxes},
			&noteRmCmd{boxes: boxes},
		},
	}, "")

	flag.Parse()

	return int(subcommands.Execute(ctx))
}
