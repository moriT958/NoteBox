package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
	"github.com/google/uuid"
)

var _ subcommands.Command = (*newCmd)(nil)

type newCmd struct {
	cfg   *config
	title string
}

// Name returns the name of the command.
func (*newCmd) Name() string { return "new" }

// Synopsis returns a short string (less than one line) describing the command.
func (*newCmd) Synopsis() string { return "create new note" }

// Usage returns a long string explaining the command and giving usage information.
func (*newCmd) Usage() string {
	return `new [-title title]:
create new note.`
}

// SetFlags adds the flags for this command to the specified set.
func (*newCmd) SetFlags(f *flag.FlagSet) {}

// Execute executes the command and returns an ExitStatus.
func (c *newCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// 引数が多すぎる時はエラーを返す
	if !validateArgs(f.Args()) {
		fmt.Fprintf(os.Stderr, "too much args. needed one.\n")
		return subcommands.ExitFailure
	}

	if len(f.Args()) > 0 {
		c.title = f.Args()[0]
	} else {
		fmt.Print("Enter Title: ")
		r := bufio.NewReader(os.Stdin)
		input, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read title: %v\n", err)
			return subcommands.ExitFailure
		}
		c.title = input
	}

	// Markdownファイルの作成
	topHeader := "# " + c.title + "\n\n"
	noteFile := filepath.Join(c.cfg.Volume, uuid.NewString()+".md")

	// TODO: すでに同じタイトルのノートが存在する場合はエラー

	fp, err := os.Create(noteFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file: %v\n", err)
		return subcommands.ExitFailure
	}
	defer fp.Close()

	fmt.Fprint(fp, topHeader)

	// TODO: lipglossでUI強化
	fmt.Printf("✅ Note Created!\nID: %d\tTitle: %s\n", 0, c.title)

	return subcommands.ExitSuccess
}

func validateArgs(args []string) bool {
	l := len(args)
	if l <= 1 {
		return true
	} else {
		return false
	}
}
