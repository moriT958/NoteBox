package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/google/subcommands"
)

// group is a command that only dispatches to its own subcommands, as in
// "nbx box list".
type group struct {
	name     string
	synopsis string
	cmds     []subcommands.Command
}

var _ subcommands.Command = (*group)(nil)

func (g *group) Name() string { return g.name }

func (g *group) Synopsis() string { return g.synopsis }

func (g *group) Usage() string {
	var b strings.Builder
	fmt.Fprintf(&b, "nbx %s <command> [args]:\n%s.\n\ncommands:\n", g.name, g.synopsis)
	for _, c := range g.cmds {
		fmt.Fprintf(&b, "  %-8s %s\n", c.Name(), c.Synopsis())
	}
	fmt.Fprintf(&b, "\nrun 'nbx help %s' or 'nbx %s help <command>' for details.\n", g.name, g.name)
	return b.String()
}

func (*group) SetFlags(f *flag.FlagSet) {}

func (g *group) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	cdr := subcommands.NewCommander(f, "nbx "+g.name)
	cdr.Register(cdr.HelpCommand(), "")
	for _, c := range g.cmds {
		cdr.Register(c, "")
	}
	return cdr.Execute(ctx, args...)
}
