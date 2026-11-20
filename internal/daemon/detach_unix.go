//go:build !windows

package daemon

import "syscall"

// detached makes the daemon its own session, so it survives the terminal
// that started it.
func detached() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
