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

type editCmd struct{}

var _ subcommands.Command = (*editCmd)(nil)

func (*editCmd) Name() string { return "edit" }

func (*editCmd) Synopsis() string { return "edit note by your editor" }

func (*editCmd) Usage() string {
	return `edit [id]:
edit note by your editor`
}

func (*editCmd) SetFlags(f *flag.FlagSet) {}

func (c *editCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {

	id, err := getIdArg(f.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	note, err := Nr.FindByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "note does't exit: %v\n", err)
		return subcommands.ExitFailure
	}

	// Noteから得たPathを指定して、editorで開く
	cmd := exec.Command(config.Editor(), note.GetFilePath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file by your editor: %v\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
