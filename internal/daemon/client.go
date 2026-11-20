package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Client returns an HTTP client that talks to the daemon over its unix
// socket. Request URLs use the host "nbx", e.g. http://nbx/api/boxes.
func Client() (*http.Client, error) {
	sock, err := SocketPath()
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", sock)
			},
		},
	}, nil
}

// Status returns the info of the running daemon, or an error when there is
// none.
func Status(ctx context.Context) (Info, error) {
	c, err := Client()
	if err != nil {
		return Info{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://nbx/api/daemon", nil)
	resp, err := c.Do(req)
	if err != nil {
		return Info{}, fmt.Errorf("the daemon is not running")
	}
	defer resp.Body.Close()
	var info Info
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return Info{}, fmt.Errorf("unexpected answer from the daemon: %w", err)
	}
	return info, nil
}

// Stop asks the running daemon to exit.
func Stop(ctx context.Context) error {
	c, err := Client()
	if err != nil {
		return err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://nbx/api/daemon/stop", nil)
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("the daemon is not running")
	}
	resp.Body.Close()
	return nil
}

// Ensure returns the running daemon, starting one in the background first
// when there is none. args are extra flags for `nbx daemon run`.
func Ensure(ctx context.Context, args ...string) (Info, error) {
	if info, err := Status(ctx); err == nil {
		return info, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return Info{}, err
	}
	dir, err := appDir()
	if err != nil {
		return Info{}, err
	}
	logf, err := os.OpenFile(filepath.Join(dir, "daemon.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return Info{}, err
	}
	defer logf.Close()

	cmd := exec.Command(exe, append([]string{"daemon", "run"}, args...)...)
	cmd.Stdout = logf
	cmd.Stderr = logf
	cmd.SysProcAttr = detached()
	if err := cmd.Start(); err != nil {
		return Info{}, fmt.Errorf("failed to start the daemon: %w", err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case err := <-exited:
			return Info{}, fmt.Errorf("the daemon exited right away (%v); see %s", err, logf.Name())
		case <-deadline:
			return Info{}, fmt.Errorf("the daemon did not start in time; see %s", logf.Name())
		case <-time.After(50 * time.Millisecond):
			if info, err := Status(ctx); err == nil {
				return info, nil
			}
		}
	}
}

// WebURL is the address that signs the browser in and opens the note at
// boxID/notePath (both optional).
func WebURL(info Info, token, boxID, notePath string) string {
	u := url.URL{Scheme: "http", Host: info.Addr, Path: "/", RawQuery: url.Values{"token": {token}}.Encode()}
	if boxID != "" {
		u.Fragment = "/b/" + boxID
		if notePath != "" {
			u.Fragment += "/" + notePath
		}
	}
	return u.String()
}
