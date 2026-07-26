package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// GetDistFS returns the embedded frontend dist filesystem.
// If the dist directory is empty (e.g. during development), returns nil.
func GetDistFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil
	}
	return sub
}
