package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/config"
	"notebox/store"
	"os"

	"github.com/google/subcommands"
)

type lsCmd struct {
	cfg   *config.Config
	store store.Store
}

var _ subcommands.Command = (*lsCmd)(nil)

func (*lsCmd) Name() string { return "ls" }

func (*lsCmd) Synopsis() string { return "list all notes" }

func (*lsCmd) Usage() string {
	return `ls:
list all notes`
}

func (*lsCmd) SetFlags(f *flag.FlagSet) {}

func (c *lsCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// StoreからNoteの一覧を取得
	notes, err := c.store.FindAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed get list of notes: %v\n", err)
	}

	// 取得したデータを標準出力に書き出す
	for i := range len(notes) {
		fmt.Printf("ID: %d\tTitle: %s\tPath: %s\n", notes[i].ID, notes[i].Title, notes[i].Path)
	}

	return subcommands.ExitSuccess
}
