package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"notebox/internal/core/box"
)

// memStore is an in-memory box.BoxStore.
type memStore struct {
	mu    sync.Mutex
	boxes map[string]box.Box
}

func (s *memStore) Set(_ context.Context, b box.Box) (*box.Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.boxes[b.ID] = b
	return &b, nil
}

func (s *memStore) GetByID(_ context.Context, id string) (*box.Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.boxes[id]; ok {
		return &b, nil
	}
	return nil, nil
}

func (s *memStore) Get(_ context.Context, f box.Filter) ([]box.Box, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []box.Box
	for _, b := range s.boxes {
		if f.Path != nil && b.Path != *f.Path {
			continue
		}
		if f.Active != nil && b.Active != *f.Active {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

func (s *memStore) Del(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.boxes, id)
	return nil
}

func newTestServer(t *testing.T) (*Server, box.Box) {
	t.Helper()
	dir := t.TempDir()
	store := &memStore{boxes: map[string]box.Box{}}
	b := box.Box{ID: "b1", Title: "Test", Path: dir, Active: true}
	store.boxes[b.ID] = b
	if err := os.WriteFile(filepath.Join(dir, "memo.md"), []byte("# memo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return New(box.NewBoxService(t.TempDir(), store)), b
}

func do(t *testing.T, h http.Handler, method, url string, body string, header map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestNoteReadWrite(t *testing.T) {
	s, b := newTestServer(t)
	h := s.API()
	url := "/api/boxes/" + b.ID + "/notes/memo.md"

	rec := do(t, h, "GET", url, "", nil)
	if rec.Code != 200 || rec.Body.String() != "# memo\n" {
		t.Fatalf("GET = %d %q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/markdown") {
		t.Errorf("Content-Type = %q", ct)
	}
	tag := rec.Header().Get("ETag")

	if rec := do(t, h, "GET", url, "", map[string]string{"If-None-Match": tag}); rec.Code != http.StatusNotModified {
		t.Errorf("GET If-None-Match = %d, want 304", rec.Code)
	}

	rec = do(t, h, "PUT", url, "# memo\n\nv2\n", map[string]string{"If-Match": tag})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT = %d %s", rec.Code, rec.Body.String())
	}
	newTag := rec.Header().Get("ETag")

	// a stale ETag is refused, and the current one is reported
	rec = do(t, h, "PUT", url, "lost update", map[string]string{"If-Match": tag})
	if rec.Code != http.StatusPreconditionFailed || rec.Header().Get("ETag") != newTag {
		t.Fatalf("stale PUT = %d etag %q, want 412 %q", rec.Code, rec.Header().Get("ETag"), newTag)
	}
	got, _ := os.ReadFile(filepath.Join(b.Path, "memo.md"))
	if string(got) != "# memo\n\nv2\n" {
		t.Errorf("file = %q", got)
	}
}

func TestNotePaths(t *testing.T) {
	s, b := newTestServer(t)
	h := s.API()
	// "../x" in the raw path is cleaned (and redirected) by ServeMux; an
	// encoded one reaches the handler and must be refused there
	for _, p := range []string{"..%2Fescape.md", "missing.md", "memo.txt"} {
		if rec := do(t, h, "GET", "/api/boxes/"+b.ID+"/notes/"+p, "", nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", p, rec.Code)
		}
	}
	if rec := do(t, h, "GET", "/api/boxes/nope/notes", "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("unknown box = %d, want 404", rec.Code)
	}
}

func TestWebAuth(t *testing.T) {
	s, b := newTestServer(t)
	static := fstest.MapFS{"index.html": {Data: []byte("<html>")}}
	h := s.Web("secret", static)
	api := "/api/boxes/" + b.ID + "/notes"
	cookie := map[string]string{"Cookie": tokenCookie + "=secret"}

	cases := []struct {
		name   string
		method string
		url    string
		header map[string]string
		host   string
		want   int
	}{
		{"no cookie", "GET", api, nil, "", http.StatusUnauthorized},
		{"wrong cookie", "GET", api, map[string]string{"Cookie": tokenCookie + "=nope"}, "", http.StatusUnauthorized},
		{"cookie", "GET", api, cookie, "", http.StatusOK},
		{"foreign host (DNS rebinding)", "GET", api, cookie, "evil.example:8765", http.StatusForbidden},
		{"cross-origin write", "POST", api, map[string]string{"Cookie": tokenCookie + "=secret", "Origin": "http://evil.example"}, "", http.StatusForbidden},
		{"token sign-in", "GET", "/?token=secret", nil, "", http.StatusSeeOther},
		{"bad token", "GET", "/?token=nope", nil, "", http.StatusUnauthorized},
		{"spa fallback", "GET", "/some/route", cookie, "", http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.url, strings.NewReader(`{"title":"x"}`))
			req.Host = "127.0.0.1:8765"
			if c.host != "" {
				req.Host = c.host
			}
			for k, v := range c.header {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, c.want, rec.Body.String())
			}
		})
	}
}
