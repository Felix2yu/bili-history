package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// GetDistFS returns the embedded frontend dist filesystem.
// dist always contains at least .gitkeep so go:embed can compile in a fresh
// checkout; a development build therefore has a filesystem but no index.html,
// which simply leaves the web UI unserved.
func GetDistFS() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil
	}
	return sub
}
