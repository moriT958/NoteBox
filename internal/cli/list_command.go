package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/subcommands"
	"github.com/mattn/go-runewidth"

	"notebox/internal/database"
)

type listCmd struct {
	verbose bool
}

var _ subcommands.Command = (*listCmd)(nil)

func (*listCmd) Name() string { return "list" }

func (*listCmd) Synopsis() string { return "list all boxes" }

func (*listCmd) Usage() string {
	return `notebox list [-v]:
show all boxes. deleted boxes are shown dimmed.
a merged box shows its primary path with a "(+N merged)" count;
pass -v to list every merged directory on its own line instead.
`
}

func (c *listCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.verbose, "v", false, "list every merged directory on its own line")
	f.BoolVar(&c.verbose, "verbose", false, "list every merged directory on its own line")
}

func (c *listCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	db, err := database.NewSQLiteDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open database:", err)
		return subcommands.ExitFailure
	}
	defer db.Close()

	boxes, err := database.NewBoxRepository(db).FindAll(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to list boxes:", err)
		return subcommands.ExitFailure
	}

	const (
		dim    = "\x1b[2m"
		header = "\x1b[1;36m"
		reset  = "\x1b[0m"
	)

	titleHeader := "Box"
	width := runewidth.StringWidth(titleHeader)
	for _, b := range boxes {
		if l := runewidth.StringWidth(b.Title); l > width {
			width = l
		}
	}

	pad := func(s string) string {
		return s + strings.Repeat(" ", width-runewidth.StringWidth(s)+2)
	}
	indent := strings.Repeat(" ", width+2)

	fmt.Println(header + pad(titleHeader) + "Path" + reset)
	for _, b := range boxes {
		wrap := func(s string) string { return s }
		if !b.Active {
			wrap = func(s string) string { return dim + s + reset }
		}

		fmt.Println(wrap(pad(b.Title) + shortenHomePath(b.Path) + mergedSuffix(c.verbose, len(b.Paths))))
		if c.verbose {
			for _, p := range b.Paths {
				fmt.Println(wrap(indent + shortenHomePath(p)))
			}
		}
	}

	return subcommands.ExitSuccess
}

// mergedSuffix reports, for the compact (non-verbose) view, how many extra
// directories are merged into a box; the verbose view lists them instead.
func mergedSuffix(verbose bool, mergedCount int) string {
	if verbose || mergedCount == 0 {
		return ""
	}
	return fmt.Sprintf(" (+%d merged)", mergedCount)
}
