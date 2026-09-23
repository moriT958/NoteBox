package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testWatchTimeout = 2 * time.Second

func waitForSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if !ok {
			t.Fatalf("channel closed unexpectedly")
		}
	case <-time.After(testWatchTimeout):
		t.Fatalf("timed out waiting for signal")
	}
}

func waitForClose(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatalf("expected channel to be closed, got a value instead")
		}
	case <-time.After(testWatchTimeout):
		t.Fatalf("timed out waiting for channel to close")
	}
}

func TestWatcher_Watch_InitialSignal(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher()
	defer w.Close()

	ch, err := w.Watch(dir)
	if err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}
	waitForSignal(t, ch)
}

func TestWatcher_Watch_DetectsFileChange(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher()
	defer w.Close()

	ch, err := w.Watch(dir)
	if err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}
	waitForSignal(t, ch) // initial

	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("# note"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	waitForSignal(t, ch)
}

func TestWatcher_Watch_DetectsChangeInNewSubdirectory(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher()
	defer w.Close()

	ch, err := w.Watch(dir)
	if err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}
	waitForSignal(t, ch) // initial

	subDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	waitForSignal(t, ch) // subdir creation itself, by which time it's already Add()-ed

	if err := os.WriteFile(filepath.Join(subDir, "note.md"), []byte("# note"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	waitForSignal(t, ch)
}

func TestWatcher_Watch_ReplacesPreviousWatch(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	w := NewWatcher()
	defer w.Close()

	chA, err := w.Watch(dirA)
	if err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}
	waitForSignal(t, chA) // initial

	chB, err := w.Watch(dirB)
	if err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}

	waitForClose(t, chA)
	waitForSignal(t, chB) // initial
}

func TestWatcher_Close_StopsWatching(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher()

	ch, err := w.Watch(dir)
	if err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}
	waitForSignal(t, ch) // initial

	if err := w.Close(); err != nil {
		t.Fatalf("unexpected err occurred: %v", err)
	}

	waitForClose(t, ch)
}
