package cli

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"notebox/internal/core/box"
	"notebox/internal/daemon"

	"github.com/google/subcommands"
)

type openCmd struct {
	boxes    *box.BoxService
	boxPath  string
	printURL bool
}

var _ subcommands.Command = (*openCmd)(nil)

func (*openCmd) Name() string     { return "open" }
func (*openCmd) Synopsis() string { return "open a box or note in the web UI" }
func (*openCmd) Usage() string {
	return `nbx open [-box <path>] [-print-url] [<note-path>]:
open the box (or the note in it) in the web UI, starting the daemon when it is
not running. the URL is always printed; with -print-url the browser is not
opened (use it over ssh: forward the port with 'ssh -L 8765:localhost:8765').
`
}

func (c *openCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.boxPath, "box", "", boxFlagUsage)
	f.BoolVar(&c.printURL, "print-url", false, "only print the URL")
}

func (c *openCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	pos, err := positional(f)
	if err != nil || len(pos) > 1 {
		f.Usage()
		return subcommands.ExitUsageError
	}

	b, err := resolveBox(ctx, c.boxes, c.boxPath)
	if err != nil {
		return fail(err)
	}
	notePath := ""
	if len(pos) == 1 {
		svc, err := notes(ctx, c.boxes, c.boxPath)
		if err != nil {
			return fail(err)
		}
		n, err := svc.GetNote(pos[0])
		if err != nil {
			return fail(err)
		}
		notePath = url.PathEscape(n.Path())
	}

	info, err := daemon.Ensure(ctx)
	if err != nil {
		return fail(err)
	}
	token, err := daemon.Token()
	if err != nil {
		return fail(err)
	}
	u := daemon.WebURL(info, token, b.ID, notePath)
	fmt.Println(u)
	if !c.printURL {
		if err := openBrowser(u); err != nil {
			return fail(fmt.Errorf("failed to open the browser: %w", err))
		}
	}
	return subcommands.ExitSuccess
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}
