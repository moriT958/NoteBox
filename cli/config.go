package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/config"
	"os"
	"os/exec"

	"github.com/google/subcommands"
)

type configCmd struct{}

var _ subcommands.Command = (*configCmd)(nil)

func (*configCmd) Name() string { return "config" }

func (*configCmd) Synopsis() string { return "edit config file" }

func (*configCmd) Usage() string {
	return `config:
open config file and you can edit`
}

func (*configCmd) SetFlags(f *flag.FlagSet) {}

func (c *configCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {

	cmd := exec.Command(config.Editor, config.ConfigFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file by your editor: %v\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
