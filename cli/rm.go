package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

type rmCmd struct{}

var _ subcommands.Command = (*rmCmd)(nil)

func (*rmCmd) Name() string { return "rm" }

func (*rmCmd) Synopsis() string { return "delete note by id" }

func (*rmCmd) Usage() string {
	return `rm [id]:
delete note and note metadata by id.
`
}

func (*rmCmd) SetFlags(f *flag.FlagSet) {}

func (c *rmCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	id, err := getIdArg(f.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	note, err := Nr.FindByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get note by id: %v\n", err)
		return subcommands.ExitFailure
	}

	if err := Nr.DeleteByID(id); err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete note: %v\n", err)
		return subcommands.ExitFailure
	}

	if err := os.Remove(note.GetFilePath()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove note file: %v\n", err)
		return subcommands.ExitFailure
	}

	fmt.Printf("🗑️ Deleted!\tID: %d\n", id)

	return subcommands.ExitSuccess
}
