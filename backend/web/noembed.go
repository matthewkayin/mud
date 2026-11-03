//go:build !embed

package web

import (
	"io/fs"
)

// Without the embed build tag the frontend is served separately (by Vite during development)
func getFrontendFiles() fs.FS {
	return nil
}
