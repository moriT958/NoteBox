package watcher

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofsnotify/fsnotify"
)

const watchOps = fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write

const debounceDelay = 100 * time.Millisecond

// Watcher recursively watches a directory tree and reports (debounced)
// whenever something inside it changes. It has no knowledge of what the
// directory contains; the caller decides what to reload.
type Watcher struct {
	mu     sync.Mutex
	fsw    *fsnotify.Watcher
	cancel context.CancelFunc
}

func NewWatcher() *Watcher {
	return &Watcher{}
}

// Watch replaces any previous watch and starts watching path and all of
// its subdirectories, including ones created after Watch is called. The
// returned channel receives a signal (debounced, non-blocking) whenever
// something changes; it's closed when Watch is called again or Close is
// called.
func (w *Watcher) Watch(path string) (<-chan struct{}, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.stopLocked()

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	dirs, err := collectDirs(path)
	if err != nil {
		fsw.Close()
		return nil, err
	}
	for _, d := range dirs {
		if err := fsw.Add(d, watchOps); err != nil {
			fsw.Close()
			return nil, err
		}
	}
	w.fsw = fsw

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	ch := make(chan struct{}, 1)
	ch <- struct{}{}

	go watchLoop(ctx, fsw, ch)
	return ch, nil
}

// Close stops watching and releases all underlying resources.
func (w *Watcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.stopLocked()
}

func (w *Watcher) stopLocked() error {
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
	if w.fsw == nil {
		return nil
	}
	err := w.fsw.Close()
	w.fsw = nil
	return err
}

func collectDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}

func watchLoop(ctx context.Context, fsw *fsnotify.Watcher, ch chan struct{}) {
	defer close(ch)

	var debounceTimer *time.Timer

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-fsw.Events:
			if !ok {
				return
			}
			// Watch newly created subdirectories too, so changes inside
			// them are picked up as well.
			if ev.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					if err := fsw.Add(ev.Name, watchOps); err != nil {
						slog.Error("watch new directory", slog.String("path", ev.Name), slog.String("error", err.Error()))
					}
				}
			}

			// Debounce rapid successive events caused by editor's atomic write.
			if debounceTimer != nil {
				debounceTimer.Stop()
				select {
				case <-debounceTimer.C:
				default:
				}
			}
			debounceTimer = time.NewTimer(debounceDelay)
		case <-timerC(debounceTimer):
			debounceTimer = nil
			select {
			case ch <- struct{}{}:
			default:
			}
		case err, ok := <-fsw.Errors:
			if !ok {
				return
			}
			slog.Error("fsnotify error", slog.String("error", err.Error()))
		}
	}
}

// timerC avoids a nil dereference (t.C) and disables the select case when
// no timer is set.
func timerC(t *time.Timer) <-chan time.Time {
	if t == nil {
		return nil
	}
	return t.C
}
