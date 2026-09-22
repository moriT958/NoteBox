package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"notebox/internal/notedeprecated"
	"os"
	"strings"

	"github.com/google/subcommands"
)

type pruneCmd struct {
	repo  notedeprecated.BoxRepository
	force bool
}

var _ subcommands.Command = (*pruneCmd)(nil)

func (*pruneCmd) Name() string { return "prune" }

func (*pruneCmd) Synopsis() string { return "remove deleted boxes and their directories" }

func (*pruneCmd) Usage() string {
	return `notebox prune [-force]:
cleanup directories of inactive boxes.
`
}

func (c *pruneCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.force, "force", false, "skip confirmation prompt")
}

func (c *pruneCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	boxes, err := c.repo.FindInactiveBoxes(ctx)
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

	for _, b := range boxes {
		if b.Path == "" {
			continue
		}
		if err := os.RemoveAll(b.Path); err != nil {
			fmt.Fprintln(os.Stderr, "failed to remove directory:", err)
			return subcommands.ExitFailure
		}
	}

	if err := c.repo.PruneBoxes(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "failed to prune boxes:", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("pruned %d box\n", len(boxes))
	return subcommands.ExitSuccess
}
