package cli

import (
	"context"
	"flag"
	"fmt"
	"notebox/internal/core/box"
	"strings"

	"github.com/google/subcommands"
)

func printBox(b box.Box) {
	fmt.Println(b.Title + "\t" + b.Path)
}

type boxAddCmd struct {
	boxes *box.BoxService
	title string
}

var _ subcommands.Command = (*boxAddCmd)(nil)

func (*boxAddCmd) Name() string { return "add" }

func (*boxAddCmd) Synopsis() string { return "add a directory as a box" }

func (*boxAddCmd) Usage() string {
	return `nbx box add [-title <title>] <path>:
register the directory at <path> as a box. the directory is created when it
doesn't exist. the title defaults to the directory name. adding the path of a
removed box brings it back.
`
}

func (c *boxAddCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.title, "title", "", "title of the box (default: directory name)")
}

func (c *boxAddCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 1)
	if !ok {
		return subcommands.ExitUsageError
	}

	b, err := c.boxes.AddBox(ctx, c.title, pos[0])
	if err != nil {
		return fail(err)
	}
	printBox(*b)
	return subcommands.ExitSuccess
}

type boxRenameCmd struct {
	boxes *box.BoxService
}

var _ subcommands.Command = (*boxRenameCmd)(nil)

func (*boxRenameCmd) Name() string { return "rename" }

func (*boxRenameCmd) Synopsis() string { return "rename a box" }

func (*boxRenameCmd) Usage() string {
	return `nbx box rename <path> <new-title>:
change the title of the box at <path>.
`
}

func (*boxRenameCmd) SetFlags(f *flag.FlagSet) {}

func (c *boxRenameCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 2)
	if !ok {
		return subcommands.ExitUsageError
	}
	if strings.TrimSpace(pos[1]) == "" {
		return fail(fmt.Errorf("title is required"))
	}

	b, err := resolveBox(ctx, c.boxes, pos[0])
	if err != nil {
		return fail(err)
	}
	renamed, err := c.boxes.RenameBox(ctx, b.ID, pos[1])
	if err != nil {
		return fail(err)
	}
	printBox(*renamed)
	return subcommands.ExitSuccess
}

type boxRelinkCmd struct {
	boxes *box.BoxService
}

var _ subcommands.Command = (*boxRelinkCmd)(nil)

func (*boxRelinkCmd) Name() string { return "relink" }

func (*boxRelinkCmd) Synopsis() string { return "point a box at another directory" }

func (*boxRelinkCmd) Usage() string {
	return `nbx box relink <path> <new-path>:
point the box at <path> to <new-path>. files are not moved.
`
}

func (*boxRelinkCmd) SetFlags(f *flag.FlagSet) {}

func (c *boxRelinkCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 2)
	if !ok {
		return subcommands.ExitUsageError
	}

	b, err := resolveBox(ctx, c.boxes, pos[0])
	if err != nil {
		return fail(err)
	}
	relinked, err := c.boxes.ChangeBoxPath(ctx, b.ID, pos[1])
	if err != nil {
		return fail(err)
	}
	printBox(*relinked)
	return subcommands.ExitSuccess
}

type boxRmCmd struct {
	boxes *box.BoxService
}

var _ subcommands.Command = (*boxRmCmd)(nil)

func (*boxRmCmd) Name() string { return "rm" }

func (*boxRmCmd) Synopsis() string { return "remove a box (files are kept)" }

func (*boxRmCmd) Usage() string {
	return `nbx box rm <path>:
remove the box at <path>. its directory is kept until 'nbx box prune'.
the default box cannot be removed.
`
}

func (*boxRmCmd) SetFlags(f *flag.FlagSet) {}

func (c *boxRmCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, ok := wantArgs(f, 1)
	if !ok {
		return subcommands.ExitUsageError
	}

	b, err := resolveBox(ctx, c.boxes, pos[0])
	if err != nil {
		return fail(err)
	}
	if err := c.boxes.RemoveBox(ctx, b.ID); err != nil {
		return fail(err)
	}
	return subcommands.ExitSuccess
}
