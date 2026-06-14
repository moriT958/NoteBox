package note

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gofsnotify/fsnotify"
)

type Registerer interface {
	Register(path string) (<-chan []Note, error)
	Unregister(path string) error
}

type FSNotifyRegisterer struct {
	*fsnotify.Watcher
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewFSNotifyRegisterer() (*FSNotifyRegisterer, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FSNotifyRegisterer{
		Watcher: w,
		cancels: make(map[string]context.CancelFunc),
	}, nil
}

var _ Registerer = (*FSNotifyRegisterer)(nil)

func (r *FSNotifyRegisterer) Register(path string) (<-chan []Note, error) {
	notes, err := LoadNoteFiles(path)
	if err != nil {
		return nil, err
	}

	if err := r.Add(path, fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Write); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	// NOTE: overwriting cancels[path] without calling the old cancel would leak the
	// previous watch goroutine. However, Register is never called twice for the same
	// path — switchBox always calls Unregister first — so this is not an issue in practice.
	r.mu.Lock()
	r.cancels[path] = cancel
	r.mu.Unlock()

	ch := make(chan []Note, 1)
	ch <- notes

	go r.watch(ctx, path, ch)
	return ch, nil
}

func (r *FSNotifyRegisterer) Unregister(path string) error {
	r.mu.Lock()
	if cancel, ok := r.cancels[path]; ok {
		cancel()
		delete(r.cancels, path)
	}
	r.mu.Unlock()

	return r.Watcher.Remove(path)
}

func (r *FSNotifyRegisterer) watch(ctx context.Context, path string, ch chan<- []Note) {
	defer close(ch)

	// Debounce rapid successive events caused by editor's atomic write
	var debounceTimer *time.Timer

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-r.Events:
			if !ok {
				return
			}
			if debounceTimer != nil {
				debounceTimer.Stop()
				// drain any already-fired tick to avoid a spurious reload
				select {
				case <-debounceTimer.C:
				default:
				}
			}
			debounceTimer = time.NewTimer(debounceDelay)
		case <-timerC(debounceTimer):
			debounceTimer = nil
			notes, err := LoadNoteFiles(path)
			if err != nil {
				slog.Error("reload notes", slog.String("error", err.Error()))
				continue
			}
			ch <- notes
		case err, ok := <-r.Errors:
			if !ok {
				return
			}
			slog.Error("fsnotify error", slog.String("error", err.Error()))
		}
	}
}

const debounceDelay = 100 * time.Millisecond

// timerC avoids nil dereference(t.C) and disables the select case when no timer is set.
func timerC(t *time.Timer) <-chan time.Time {
	if t == nil {
		return nil
	}
	return t.C
}
