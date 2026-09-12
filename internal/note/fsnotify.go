package note

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gofsnotify/fsnotify"
)

type Registerer interface {
	// Register watches all of the given paths as a single unit (e.g. the
	// directories merged into one box) and returns a channel of their
	// combined notes.
	Register(paths []string) (<-chan []Note, error)
	Unregister(paths []string) error
}

type FSNotifyRegisterer struct {
	*fsnotify.Watcher
	mu sync.Mutex
	// cancel stops the watch goroutine of the currently registered path
	// group. Only one group (i.e. one box's directories) is ever registered
	// at a time — switchBox and rebindCurrentBox always call Unregister
	// before Register — so a single cancel func is enough to track it.
	cancel context.CancelFunc
}

func NewFSNotifyRegisterer() (*FSNotifyRegisterer, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FSNotifyRegisterer{
		Watcher: w,
	}, nil
}

var _ Registerer = (*FSNotifyRegisterer)(nil)

func (r *FSNotifyRegisterer) Register(paths []string) (<-chan []Note, error) {
	notes, err := LoadNoteFiles(paths...)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if err := r.Add(path, fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Write); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.cancel = cancel
	r.mu.Unlock()

	ch := make(chan []Note, 1)
	ch <- notes

	go r.watch(ctx, paths, ch)
	return ch, nil
}

func (r *FSNotifyRegisterer) Unregister(paths []string) error {
	r.mu.Lock()
	cancel := r.cancel
	r.cancel = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	var firstErr error
	for _, path := range paths {
		if err := r.Watcher.Remove(path); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *FSNotifyRegisterer) watch(ctx context.Context, paths []string, ch chan<- []Note) {
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
			notes, err := LoadNoteFiles(paths...)
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
