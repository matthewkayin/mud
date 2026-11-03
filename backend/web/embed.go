//go:build embed

package web

import (
	"embed"
	"io/fs"
	"log"
)

// Built by `npm run build` in frontend/, which writes to backend/web/dist
//go:embed all:dist
var distFiles embed.FS

func getFrontendFiles() fs.FS {
	frontendFiles, err := fs.Sub(distFiles, "dist")
	if err != nil {
		log.Fatalf("Failed to load embedded frontend: %s", err.Error())
	}
	return frontendFiles
}
