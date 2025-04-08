package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"notebox/models"
	"os"
	"strings"
	"time"

	"github.com/google/subcommands"
)

var _ subcommands.Command = (*newCmd)(nil)

type newCmd struct{}

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
func (c *newCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	title, err := getTitleArg(f.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return subcommands.ExitFailure
	}

	// Noteのメタデータを保存
	note := new(models.Note)
	note.SetTitle(title)
	note.SetCreatedAt(time.Now())

	id, err := Nr.Save(*note)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save note: %v\n", err)
	}

	// Markdownファイルの作成
	topHeader := "# " + title + "\n\n"
	fp, err := os.Create(note.GetFilePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file: %v\n", err)
		return subcommands.ExitFailure
	}
	defer fp.Close()
	fmt.Fprint(fp, topHeader)
	fmt.Printf("✅ Note Created!\nID: %d\tTitle: %s\n", id, note.Title())

	return subcommands.ExitSuccess
}

func getTitleArg(args []string) (string, error) {
	var title string

	if !validateArgs(args) {
		return "", errors.New("too much args. needed one.")
	}

	if len(args) > 0 {
		title = args[0]
	} else {
		fmt.Print("Enter Title: ")
		r := bufio.NewReader(os.Stdin)
		input, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		title = strings.TrimSuffix(input, "\n")
	}

	return title, nil
}
