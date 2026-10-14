package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/internal/core/box"

	"github.com/google/subcommands"
)

type boxListCmd struct {
	boxes *box.BoxService
}

var _ subcommands.Command = (*boxListCmd)(nil)

func (*boxListCmd) Name() string { return "list" }

func (*boxListCmd) Synopsis() string { return "list all boxes" }

func (*boxListCmd) Usage() string {
	return `nbx box list:
show all boxes. on a terminal it is a table and removed boxes are dimmed.
otherwise it is one box per line as "<title>	<path>".
`
}

func (*boxListCmd) SetFlags(f *flag.FlagSet) {}

func (c *boxListCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}

	boxes, err := c.boxes.GetBoxes(ctx)
	if err != nil {
		return fail(fmt.Errorf("failed to list boxes: %w", err))
	}

	rows := make([][]string, len(boxes))
	dimmed := make([]bool, len(boxes))
	for i, b := range boxes {
		rows[i] = []string{b.Title, b.Path}
		dimmed[i] = !b.Active
	}
	printTable([]string{"Box", "Path"}, rows, dimmed)

	return subcommands.ExitSuccess
}
