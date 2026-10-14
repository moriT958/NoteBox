package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/internal/core/box"
	"notebox/internal/core/note"
	"strings"

	"github.com/google/subcommands"
)

const boxFlagUsage = "path of the box (default: the default box)"

func printNote(n note.Note) {
	fmt.Println(n.Title() + "\t" + n.Path())
}

// notes returns the note service of the box given by -box.
func notes(ctx context.Context, boxes *box.BoxService, boxPath string) (*note.NoteService, error) {
	b, err := resolveBox(ctx, boxes, boxPath)
	if err != nil {
		return nil, err
	}
	return note.NewNoteService(b), nil
}

type noteListCmd struct {
	boxes   *box.BoxService
	boxPath string
}

var _ subcommands.Command = (*noteListCmd)(nil)

func (*noteListCmd) Name() string { return "list" }

func (*noteListCmd) Synopsis() string { return "list notes in a box" }

func (*noteListCmd) Usage() string {
	return `nbx note list [-box <path>]:
show the notes in the box, one per line as "<title>	<note-path>".
<note-path> is relative to the box and is what the other note commands take.
`
}

func (c *noteListCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.boxPath, "box", "", boxFlagUsage)
}

func (c *noteListCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}

	svc, err := notes(ctx, c.boxes, c.boxPath)
	if err != nil {
		return fail(err)
	}
	list, err := svc.GetNotes()
	if err != nil {
		return fail(err)
	}
	for _, n := range list {
		printNote(n)
	}
	return subcommands.ExitSuccess
}

type noteNewCmd struct {
	boxes   *box.BoxService
	boxPath string
}

var _ subcommands.Command = (*noteNewCmd)(nil)

func (*noteNewCmd) Name() string { return "new" }

func (*noteNewCmd) Synopsis() string { return "create a note" }

func (*noteNewCmd) Usage() string {
	return `nbx note new [-box <path>] <title>:
create a note in the box and show it as "<title>	<note-path>".
`
}

func (c *noteNewCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.boxPath, "box", "", boxFlagUsage)
}

func (c *noteNewCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 1)
	if !ok {
		return subcommands.ExitUsageError
	}
	if strings.TrimSpace(pos[0]) == "" {
		return fail(fmt.Errorf("title is required"))
	}

	svc, err := notes(ctx, c.boxes, c.boxPath)
	if err != nil {
		return fail(err)
	}
	n, err := svc.CreateNote(pos[0])
	if err != nil {
		return fail(err)
	}
	printNote(n)
	return subcommands.ExitSuccess
}

type noteRenameCmd struct {
	boxes   *box.BoxService
	boxPath string
}

var _ subcommands.Command = (*noteRenameCmd)(nil)

func (*noteRenameCmd) Name() string { return "rename" }

func (*noteRenameCmd) Synopsis() string { return "rename a note" }

func (*noteRenameCmd) Usage() string {
	return `nbx note rename [-box <path>] <note-path> <new-title>:
rename the file of the note and show it as "<title>	<note-path>".
the heading inside the note is not changed.
`
}

func (c *noteRenameCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.boxPath, "box", "", boxFlagUsage)
}

func (c *noteRenameCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 2)
	if !ok {
		return subcommands.ExitUsageError
	}
	if strings.TrimSpace(pos[1]) == "" {
		return fail(fmt.Errorf("title is required"))
	}

	svc, err := notes(ctx, c.boxes, c.boxPath)
	if err != nil {
		return fail(err)
	}
	n, err := svc.GetNote(pos[0])
	if err != nil {
		return fail(err)
	}
	renamed, err := svc.RenameNote(n, pos[1])
	if err != nil {
		return fail(err)
	}
	printNote(renamed)
	return subcommands.ExitSuccess
}

type noteViewCmd struct {
	boxes   *box.BoxService
	boxPath string
}

var _ subcommands.Command = (*noteViewCmd)(nil)

func (*noteViewCmd) Name() string { return "view" }

func (*noteViewCmd) Synopsis() string { return "print a note" }

func (*noteViewCmd) Usage() string {
	return `nbx note view [-box <path>] <note-path>:
print the markdown of the note as it is.
`
}

func (c *noteViewCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.boxPath, "box", "", boxFlagUsage)
}

func (c *noteViewCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 1)
	if !ok {
		return subcommands.ExitUsageError
	}

	svc, err := notes(ctx, c.boxes, c.boxPath)
	if err != nil {
		return fail(err)
	}
	n, err := svc.GetNote(pos[0])
	if err != nil {
		return fail(err)
	}
	content, err := svc.GetNoteContent(n)
	if err != nil {
		return fail(err)
	}
	fmt.Print(content)
	return subcommands.ExitSuccess
}

type noteRmCmd struct {
	boxes   *box.BoxService
	boxPath string
}

var _ subcommands.Command = (*noteRmCmd)(nil)

func (*noteRmCmd) Name() string { return "rm" }

func (*noteRmCmd) Synopsis() string { return "delete a note" }

func (*noteRmCmd) Usage() string {
	return `nbx note rm [-box <path>] <note-path>:
delete the file of the note. it cannot be undone.
`
}

func (c *noteRmCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.boxPath, "box", "", boxFlagUsage)
}

func (c *noteRmCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 1)
	if !ok {
		return subcommands.ExitUsageError
	}

	svc, err := notes(ctx, c.boxes, c.boxPath)
	if err != nil {
		return fail(err)
	}
	n, err := svc.GetNote(pos[0])
	if err != nil {
		return fail(err)
	}
	if err := svc.RemoveNote(n); err != nil {
		return fail(err)
	}
	return subcommands.ExitSuccess
}
