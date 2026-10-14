package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

// positional returns the positional arguments of f. Unlike flag.Parse alone it
// also accepts flags after them, so "nbx note new title -box path" works.
func positional(f *flag.FlagSet) ([]string, error) {
	var pos []string
	for f.NArg() > 0 {
		pos = append(pos, f.Arg(0))
		if err := f.Parse(f.Args()[1:]); err != nil {
			return nil, err
		}
	}
	return pos, nil
}

// wantArgs returns the positional arguments of f when there are exactly n of
// them, otherwise it prints the usage of the command and reports false.
func wantArgs(f *flag.FlagSet, n int) ([]string, bool) {
	args, err := positional(f)
	if err != nil {
		return nil, false
	}
	if len(args) != n {
		fmt.Fprintf(os.Stderr, "expected %d argument(s), got %d\n", n, len(args))
		f.Usage()
		return nil, false
	}
	return args, true
}

// fail prints err to stderr and returns the failure exit status.
func fail(err error) subcommands.ExitStatus {
	fmt.Fprintln(os.Stderr, err)
	return subcommands.ExitFailure
}

// isTerminal reports whether f is a terminal, so that colors are only used
// when a person is reading.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
