package note

import (
	"log/slog"
	"time"

	"github.com/gofsnotify/fsnotify"
)

type Registerer interface {
	Register(path string) (<-chan []Note, error)
}

type FSNotifyRegisterer struct {
	*fsnotify.Watcher
}

func NewFSNotifyRegisterer() (*FSNotifyRegisterer, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FSNotifyRegisterer{w}, nil
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

	ch := make(chan []Note, 1)
	ch <- notes

	go r.watch(path, ch)
	return ch, nil
}

func (r *FSNotifyRegisterer) watch(path string, ch chan<- []Note) {
	defer close(ch)

	// Debounce rapid successive events caused by editor's atomic write
	var debounceTimer *time.Timer

	for {
		select {
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
