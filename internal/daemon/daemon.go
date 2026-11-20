// Package daemon runs the NoteBox server as a long-lived process and lets
// other nbx processes find, start and stop it.
//
// The daemon listens on two sockets that share one API:
//   - a unix socket (~/.notebox/nbx.sock, mode 0600) for the CLI and TUI,
//     trusted by file permissions,
//   - a TCP address on localhost for the web UI, guarded by a token.
package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"notebox/internal/config"
	"notebox/internal/core/box"
	"notebox/internal/server"
)

// DefaultAddr is where the web UI listens unless told otherwise. A fixed port
// keeps bookmarks and `ssh -L` forwards stable.
const DefaultAddr = "127.0.0.1:8765"

const (
	socketName = "nbx.sock"
	tokenName  = "web_token"
)

// Info describes a running daemon.
type Info struct {
	PID  int    `json:"pid"`
	Addr string `json:"addr"` // web UI address
}

func appDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, config.AppDirName), nil
}

// SocketPath is the unix socket of the daemon.
func SocketPath() (string, error) {
	dir, err := appDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, socketName), nil
}

// Token returns the web UI token, creating it the first time. It is kept
// across restarts so the browser stays signed in.
func Token() (string, error) {
	dir, err := appDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, tokenName)
	if b, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(b))) > 0 {
		return strings.TrimSpace(string(b)), nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	tok := hex.EncodeToString(buf)
	if err := os.WriteFile(path, []byte(tok+"\n"), 0600); err != nil {
		return "", err
	}
	return tok, nil
}

// Run serves until ctx is done or a client asks it to stop.
func Run(ctx context.Context, boxes *box.BoxService, addr string, static fs.FS) error {
	sockPath, err := SocketPath()
	if err != nil {
		return err
	}
	if _, err := Status(ctx); err == nil {
		return errors.New("the daemon is already running")
	}
	token, err := Token()
	if err != nil {
		return fmt.Errorf("failed to prepare the web token: %w", err)
	}

	// a socket file left by a daemon that crashed
	_ = os.Remove(sockPath)
	unixLn, err := net.Listen("unix", sockPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", sockPath, err)
	}
	defer os.Remove(sockPath)
	if err := os.Chmod(sockPath, 0600); err != nil {
		unixLn.Close()
		return err
	}
	webLn, err := net.Listen("tcp", addr)
	if err != nil {
		unixLn.Close()
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	ctx, stop := context.WithCancel(ctx)
	defer stop()

	srv := server.New(boxes)
	go srv.Run(ctx)

	info := Info{PID: os.Getpid(), Addr: webLn.Addr().String()}
	local := srv.API()
	local.HandleFunc("GET /api/daemon", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})
	local.HandleFunc("POST /api/daemon/stop", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		stop()
	})

	servers := []*http.Server{
		{Handler: local},
		{Handler: srv.Web(token, static)},
	}
	errc := make(chan error, 2)
	go func() { errc <- servers[0].Serve(unixLn) }()
	go func() { errc <- servers[1].Serve(webLn) }()
	slog.Info("daemon started", "pid", info.PID, "socket", sockPath, "addr", info.Addr)

	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-errc:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, s := range servers {
		// SSE streams never end on their own; don't wait for them long
		if err := s.Shutdown(shutdownCtx); err != nil {
			s.Close()
		}
	}
	slog.Info("daemon stopped")
	if errors.Is(serveErr, http.ErrServerClosed) {
		return nil
	}
	return serveErr
}
