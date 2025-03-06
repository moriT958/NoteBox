package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"notebox/config"
	"notebox/store"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/subcommands"
)

type editCmd struct {
	cfg   *config.Config
	store store.Store
}

var _ subcommands.Command = (*editCmd)(nil)

func (*editCmd) Name() string { return "edit" }

func (*editCmd) Synopsis() string { return "edit note by your editor" }

func (*editCmd) Usage() string {
	return `edit [id]:
edit note by your editor`
}

func (*editCmd) SetFlags(f *flag.FlagSet) {}

func (c *editCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {

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

	// 引数のIDからNoteを取得する
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to convert id arg to integer: %v\n", err)
	}
	note, err := c.store.FindByID(id)

	// Noteから得たPathを指定して、vimで開く
	cmd := exec.Command(c.cfg.Editor, fmt.Sprintf("%s", note.Path))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file by your editor: %v\n", err)
	}

	return subcommands.ExitSuccess
}
