package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"notebox/internal/core/box"
	"os"
	"strings"

	"github.com/google/subcommands"
)

type boxPruneCmd struct {
	boxes *box.BoxService
	force bool
}

var _ subcommands.Command = (*boxPruneCmd)(nil)

func (*boxPruneCmd) Name() string { return "prune" }

func (*boxPruneCmd) Synopsis() string { return "remove deleted boxes and their directories" }

func (*boxPruneCmd) Usage() string {
	return `nbx box prune [-force]:
cleanup directories of inactive boxes.
`
}

func (c *boxPruneCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.force, "force", false, "skip confirmation prompt")
}

func (c *boxPruneCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}

	boxes, err := c.boxes.GetInactiveBoxes(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to find deleted boxes:", err)
		return subcommands.ExitFailure
	}

	if len(boxes) == 0 {
		fmt.Println("no boxes to prune")
		return subcommands.ExitSuccess
	}

	fmt.Println("the following directories will be removed:")
	for _, b := range boxes {
		fmt.Printf("  %s\t%s\n", b.Title, shortenHomePath(b.Path))
	}

	if !c.force {
		fmt.Print("continue? [y/N]: ")
		answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to read input:", err)
			return subcommands.ExitFailure
		}
		if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
			fmt.Println("aborted")
			return subcommands.ExitSuccess
		}
	}

	if err := c.boxes.PruneBoxes(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "failed to prune boxes:", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("pruned %d box\n", len(boxes))
	return subcommands.ExitSuccess
}
