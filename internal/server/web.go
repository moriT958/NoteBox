package server

import (
	"crypto/subtle"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"
)

const tokenCookie = "nbx_token"

// Web returns the handler of the web UI listener: the API plus the static
// files of the SPA, behind a token.
//
// The listener is bound to localhost, but any web page the user visits can
// still send requests to it, so:
//   - the Host header must be a loopback name (blocks DNS rebinding),
//   - every request must carry the token cookie, which is set by visiting
//     /?token=... once (that URL is printed by `nbx open`),
//   - the cookie is SameSite=Strict and writes must come from our origin
//     (blocks CSRF).
func (s *Server) Web(token string, static fs.FS) http.Handler {
	api := s.API()
	files := spaHandler(static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !loopbackHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}

		if q := r.URL.Query().Get("token"); q != "" {
			if !equal(q, token) {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     tokenCookie,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Expires:  time.Now().AddDate(1, 0, 0),
			})
			u := *r.URL
			q := u.Query()
			q.Del("token")
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.RequestURI(), http.StatusSeeOther)
			return
		}

		c, err := r.Cookie(tokenCookie)
		if err != nil || !equal(c.Value, token) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not signed in. Run `nbx open` to open NoteBox in the browser.\n"))
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if o := r.Header.Get("Origin"); o != "" && !sameOrigin(o, r.Host) {
				http.Error(w, "cross-origin request", http.StatusForbidden)
				return
			}
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		files.ServeHTTP(w, r)
	})
}

// spaHandler serves the built files, falling back to index.html for client
// side routes.
func spaHandler(static fs.FS) http.Handler {
	fileServer := http.FileServerFS(static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if _, err := fs.Stat(static, name); err == nil {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(static, "index.html")
		if err != nil {
			http.Error(w, "the web UI is not built into this binary (run `task web:build`)", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(index)
	})
}

func loopbackHost(hostport string) bool {
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = hostport
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func sameOrigin(origin, host string) bool {
	return origin == "http://"+host || origin == "https://"+host
}

func equal(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
