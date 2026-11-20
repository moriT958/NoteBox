// Package server is the HTTP API of NoteBox. The same handler serves the
// daemon's unix socket (CLI, TUI) and the local web UI.
//
// Note content goes over the wire as plain markdown (text/markdown); only
// metadata is JSON. Rendering is the client's job.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"notebox/internal/core/box"
	"notebox/internal/core/note"
)

// maxNoteSize caps the body of a note write.
const maxNoteSize = 16 << 20

type Server struct {
	boxes  *box.BoxService
	events *Hub
}

func New(boxes *box.BoxService) *Server {
	return &Server{boxes: boxes, events: NewHub(boxes)}
}

// Run keeps the event hub running until ctx is done.
func (s *Server) Run(ctx context.Context) { s.events.Run(ctx) }

// API returns the handler of every /api/ endpoint. It does no authentication;
// the listener in front of it decides who may call it.
func (s *Server) API() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/boxes", s.listBoxes)
	mux.HandleFunc("GET /api/boxes/{box}/notes", s.listNotes)
	mux.HandleFunc("POST /api/boxes/{box}/notes", s.createNote)
	mux.HandleFunc("GET /api/boxes/{box}/notes/{path...}", s.readNote)
	mux.HandleFunc("PUT /api/boxes/{box}/notes/{path...}", s.writeNote)
	mux.HandleFunc("PATCH /api/boxes/{box}/notes/{path...}", s.renameNote)
	mux.HandleFunc("DELETE /api/boxes/{box}/notes/{path...}", s.removeNote)
	mux.HandleFunc("GET /api/events", s.events.ServeHTTP)
	return mux
}

type boxJSON struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type noteJSON struct {
	Title string    `json:"title"`
	Path  string    `json:"path"`
	MTime time.Time `json:"mtime,omitzero"`
}

func (s *Server) listBoxes(w http.ResponseWriter, r *http.Request) {
	// make sure the default box exists, as the TUI and CLI do
	if _, err := s.boxes.DefaultBox(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	list, err := s.boxes.GetActiveBoxes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]boxJSON, len(list))
	for i, b := range list {
		out[i] = boxJSON{ID: b.ID, Title: b.Title, Path: b.Path}
	}
	writeJSON(w, http.StatusOK, out)
}

// noteService returns the notes of the box in the request path, or writes a
// 404 and returns nil.
func (s *Server) noteService(w http.ResponseWriter, r *http.Request) *note.NoteService {
	b, err := s.boxes.GetBox(r.Context(), r.PathValue("box"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return nil
	}
	if b == nil || !b.Active {
		writeError(w, http.StatusNotFound, errors.New("box not found"))
		return nil
	}
	return note.NewNoteService(*b)
}

// target returns the note in the request path, or writes an error and
// reports false.
func (s *Server) target(w http.ResponseWriter, r *http.Request) (*note.NoteService, note.Note, bool) {
	svc := s.noteService(w, r)
	if svc == nil {
		return nil, note.Note{}, false
	}
	n, err := svc.GetNote(r.PathValue("path"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return nil, note.Note{}, false
	}
	return svc, n, true
}

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	svc := s.noteService(w, r)
	if svc == nil {
		return
	}
	list, err := svc.GetNotes()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]noteJSON, len(list))
	for i, n := range list {
		out[i] = noteJSON{Title: n.Title(), Path: n.Path()}
		if info, err := os.Stat(svc.AbsPath(n)); err == nil {
			out[i].MTime = info.ModTime()
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) createNote(w http.ResponseWriter, r *http.Request) {
	svc := s.noteService(w, r)
	if svc == nil {
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	n, err := svc.CreateNote(req.Title)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusCreated, noteJSON{Title: n.Title(), Path: n.Path()})
}

func (s *Server) readNote(w http.ResponseWriter, r *http.Request) {
	svc, n, ok := s.target(w, r)
	if !ok {
		return
	}
	content, tag, err := svc.ReadNote(n)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("ETag", quote(tag))
	w.Header().Set("Cache-Control", "no-store")
	if match := r.Header.Get("If-None-Match"); match != "" && unquote(match) == tag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write(content)
}

// writeNote replaces the content of a note with the request body. With
// If-Match, a stale ETag is answered with 412 and the current ETag.
func (s *Server) writeNote(w http.ResponseWriter, r *http.Request) {
	svc, n, ok := s.target(w, r)
	if !ok {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxNoteSize))
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, err)
		return
	}
	tag, err := svc.WriteNoteContent(n, body, unquote(r.Header.Get("If-Match")))
	if errors.Is(err, note.ErrConflict) {
		w.Header().Set("ETag", quote(tag))
		writeError(w, http.StatusPreconditionFailed, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("ETag", quote(tag))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) renameNote(w http.ResponseWriter, r *http.Request) {
	svc, n, ok := s.target(w, r)
	if !ok {
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	renamed, err := svc.RenameNote(n, req.Title)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, noteJSON{Title: renamed.Title(), Path: renamed.Path()})
}

func (s *Server) removeNote(w http.ResponseWriter, r *http.Request) {
	svc, n, ok := s.target(w, r)
	if !ok {
		return
	}
	if err := svc.RemoveNote(n); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func quote(tag string) string { return `"` + tag + `"` }

func unquote(tag string) string {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "W/")
	return strings.Trim(tag, `"`)
}
