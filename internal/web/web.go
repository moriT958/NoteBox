// Package web holds the built web UI (see /web), embedded into the binary.
// `task web:build` writes the build into dist/.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// FS returns the files of the web UI. It is empty but valid when the UI was
// not built.
func FS() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
