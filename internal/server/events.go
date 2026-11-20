package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"notebox/internal/core/box"
	"notebox/internal/watcher"
)

// resyncInterval is how often the hub checks for boxes added or removed by
// another process (e.g. the CLI running without the daemon).
const resyncInterval = 3 * time.Second

// Event is sent to clients over SSE. Clients refetch what the event names;
// events carry no content.
type Event struct {
	Type string `json:"type"`          // "notes" (something in a box changed) or "boxes"
	Box  string `json:"box,omitempty"` // box ID, for "notes"
}

// Hub watches every active box and fans file changes out to SSE clients.
type Hub struct {
	boxes *box.BoxService

	mu       sync.Mutex
	subs     map[chan Event]struct{}
	watchers map[string]*boxWatch // by box ID
}

type boxWatch struct {
	path string
	w    *watcher.Watcher
}

func NewHub(boxes *box.BoxService) *Hub {
	return &Hub{
		boxes:    boxes,
		subs:     make(map[chan Event]struct{}),
		watchers: make(map[string]*boxWatch),
	}
}

// Run keeps the set of watched boxes in sync until ctx is done.
func (h *Hub) Run(ctx context.Context) {
	t := time.NewTicker(resyncInterval)
	defer t.Stop()
	h.resync(ctx)
	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			for id, bw := range h.watchers {
				bw.w.Close()
				delete(h.watchers, id)
			}
			h.mu.Unlock()
			return
		case <-t.C:
			h.resync(ctx)
		}
	}
}

func (h *Hub) resync(ctx context.Context) {
	list, err := h.boxes.GetActiveBoxes(ctx)
	if err != nil {
		slog.Error("events: failed to list boxes", "error", err)
		return
	}

	h.mu.Lock()
	changed := false
	seen := make(map[string]bool, len(list))
	for _, b := range list {
		seen[b.ID] = true
		if bw, ok := h.watchers[b.ID]; ok && bw.path == b.Path {
			continue
		} else if ok {
			bw.w.Close()
		}
		w := watcher.NewWatcher()
		ch, err := w.Watch(b.Path)
		if err != nil {
			slog.Error("events: failed to watch box", "box", b.ID, "error", err)
			continue
		}
		h.watchers[b.ID] = &boxWatch{path: b.Path, w: w}
		changed = true
		go h.forward(b.ID, ch)
	}
	for id, bw := range h.watchers {
		if !seen[id] {
			bw.w.Close()
			delete(h.watchers, id)
			changed = true
		}
	}
	h.mu.Unlock()

	if changed {
		h.publish(Event{Type: "boxes"})
	}
}

func (h *Hub) forward(boxID string, ch <-chan struct{}) {
	<-ch // the watcher signals once right away
	for range ch {
		h.publish(Event{Type: "notes", Box: boxID})
	}
}

func (h *Hub) publish(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for sub := range h.subs {
		select {
		case sub <- ev:
		default: // slow client; it will catch up on the next event
		}
	}
}

// ServeHTTP streams events as server-sent events.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")

	sub := make(chan Event, 16)
	h.mu.Lock()
	h.subs[sub] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.subs, sub)
		h.mu.Unlock()
	}()

	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
		case ev := <-sub:
			data, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		flusher.Flush()
	}
}
