package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"notebox/internal/core/box"
	"notebox/internal/daemon"
	"notebox/internal/web"

	"github.com/google/subcommands"
)

const addrFlagUsage = "address of the web UI (localhost only)"

type daemonRunCmd struct {
	boxes *box.BoxService
	addr  string
}

var _ subcommands.Command = (*daemonRunCmd)(nil)

func (*daemonRunCmd) Name() string     { return "run" }
func (*daemonRunCmd) Synopsis() string { return "run the daemon in the foreground" }
func (*daemonRunCmd) Usage() string {
	return `nbx daemon run [-addr <host:port>]:
run the daemon in the foreground until it is interrupted or stopped.
`
}

func (c *daemonRunCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.addr, "addr", daemon.DefaultAddr, addrFlagUsage)
}

func (c *daemonRunCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := daemon.Run(ctx, c.boxes, c.addr, web.FS()); err != nil {
		return fail(err)
	}
	return subcommands.ExitSuccess
}

type daemonStartCmd struct{ addr string }

var _ subcommands.Command = (*daemonStartCmd)(nil)

func (*daemonStartCmd) Name() string     { return "start" }
func (*daemonStartCmd) Synopsis() string { return "start the daemon in the background" }
func (*daemonStartCmd) Usage() string {
	return `nbx daemon start [-addr <host:port>]:
start the daemon in the background unless it is running, and show it as
"<pid>	<web-address>".
`
}

func (c *daemonStartCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.addr, "addr", daemon.DefaultAddr, addrFlagUsage)
}

func (c *daemonStartCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}
	info, err := daemon.Ensure(ctx, "-addr", c.addr)
	if err != nil {
		return fail(err)
	}
	fmt.Printf("%d\t%s\n", info.PID, info.Addr)
	return subcommands.ExitSuccess
}

type daemonStopCmd struct{}

var _ subcommands.Command = (*daemonStopCmd)(nil)

func (*daemonStopCmd) Name() string             { return "stop" }
func (*daemonStopCmd) Synopsis() string         { return "stop the daemon" }
func (*daemonStopCmd) Usage() string            { return "nbx daemon stop:\nstop the running daemon.\n" }
func (*daemonStopCmd) SetFlags(f *flag.FlagSet) {}

func (c *daemonStopCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}
	if err := daemon.Stop(ctx); err != nil {
		return fail(err)
	}
	return subcommands.ExitSuccess
}

type daemonStatusCmd struct{}

var _ subcommands.Command = (*daemonStatusCmd)(nil)

func (*daemonStatusCmd) Name() string     { return "status" }
func (*daemonStatusCmd) Synopsis() string { return "show whether the daemon is running" }
func (*daemonStatusCmd) Usage() string {
	return `nbx daemon status:
show the running daemon as "<pid>	<web-address>". exits with 1 when it is
not running.
`
}
func (*daemonStatusCmd) SetFlags(f *flag.FlagSet) {}

func (c *daemonStatusCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	if _, ok := wantArgs(f, 0); !ok {
		return subcommands.ExitUsageError
	}
	info, err := daemon.Status(ctx)
	if err != nil {
		return fail(err)
	}
	fmt.Printf("%d\t%s\n", info.PID, info.Addr)
	return subcommands.ExitSuccess
}
