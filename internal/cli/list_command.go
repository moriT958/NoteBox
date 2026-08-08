package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/google/subcommands"

	"notebox/internal/database"
)

type listCmd struct{}

var _ subcommands.Command = (*listCmd)(nil)

func (*listCmd) Name() string { return "list" }

func (*listCmd) Synopsis() string { return "list all boxes" }

func (*listCmd) Usage() string {
	return `notebox list:
show all boxes. deleted boxes are shown dimmed.
`
}

func (*listCmd) SetFlags(f *flag.FlagSet) {}

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
	width := utf8.RuneCountInString(titleHeader)
	for _, b := range boxes {
		if l := utf8.RuneCountInString(b.Title); l > width {
			width = l
		}
	}

	pad := func(s string) string {
		return s + strings.Repeat(" ", width-utf8.RuneCountInString(s)+2)
	}

	fmt.Println(header + pad(titleHeader) + "Path" + reset)
	home, _ := os.UserHomeDir()
	for _, b := range boxes {
		path := b.Path
		if home != "" && strings.HasPrefix(path, home) {
			path = "~" + strings.TrimPrefix(path, home)
		}
		line := pad(b.Title) + path
		if !b.Active {
			line = dim + line + reset
		}
		fmt.Println(line)
	}

	return subcommands.ExitSuccess
}
