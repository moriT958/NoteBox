package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

type lsCmd struct{}

var _ subcommands.Command = (*lsCmd)(nil)

func (*lsCmd) Name() string { return "ls" }

func (*lsCmd) Synopsis() string { return "list all notes" }

func (*lsCmd) Usage() string {
	return `ls:
list all notes`
}

func (*lsCmd) SetFlags(f *flag.FlagSet) {}

func (c *lsCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	// StoreからNoteの一覧を取得
	notes, err := Nr.FindAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get note list: %v\n", err)
		return subcommands.ExitFailure
	}

	// 取得したデータを標準出力に書き出す
	for i := range len(notes) {
		fmt.Printf("ID: %d\tTitle: %s\n", notes[i].ID, notes[i].Title)
	}

	return subcommands.ExitSuccess
}
