package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

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

	// 引数で取得するID
	var idStr string

	// 引数が多すぎる時はエラーを返す
	if !validateArgs(f.Args()) {
		fmt.Fprintf(os.Stderr, "too much args. needed one.\n")
		return subcommands.ExitFailure
	}

	if len(f.Args()) > 0 {
		idStr = f.Args()[0]
	} else {
		fmt.Print("Enter ID: ")
		r := bufio.NewReader(os.Stdin)
		input, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read ID: %v\n", err)
			return subcommands.ExitFailure
		}
		idStr = input
	}

	// 引数のIDからNoteを削除する
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to convert id arg to integer: %v\n", err)
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
