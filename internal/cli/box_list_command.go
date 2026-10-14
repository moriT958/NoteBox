package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/internal/core/box"
	"os"

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
show all boxes, one per line as "<title>	<path>". removed boxes are
shown dimmed when the output is a terminal.
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

	const (
		dim   = "\x1b[2m"
		reset = "\x1b[0m"
	)
	color := isTerminal(os.Stdout)

	for _, b := range boxes {
		line := b.Title + "\t" + b.Path
		if !b.Active && color {
			line = dim + line + reset
		}
		fmt.Println(line)
	}

	return subcommands.ExitSuccess
}
