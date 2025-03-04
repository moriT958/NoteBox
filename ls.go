package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

type lsCmd struct {
	cfg   *config
	notes []note
}

var _ subcommands.Command = (*lsCmd)(nil)

func (*lsCmd) Name() string { return "" }

func (*lsCmd) Synopsis() string { return "" }

func (*lsCmd) Usage() string {
	return ``
}

func (*lsCmd) SetFlags(f *flag.FlagSet) {}

func (c *lsCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	return subcommands.ExitSuccess
}
