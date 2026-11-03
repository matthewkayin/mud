package web

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Returns a handler that serves the built frontend, or nil if the frontend isn't embedded (see embed.go).
// Paths that aren't files get index.html, so that the client side router handles them.
func Handler() http.Handler {
	frontendFiles := getFrontendFiles()
	if frontendFiles == nil {
		return nil
	}

	fileServer := http.FileServerFS(frontendFiles)
	return http.HandlerFunc(func (writer http.ResponseWriter, request *http.Request) {
		filePath := strings.TrimPrefix(path.Clean("/" + request.URL.Path), "/")
		if filePath == "" {
			filePath = "."
		}

		fileInfo, err := fs.Stat(frontendFiles, filePath)
		if err != nil || fileInfo.IsDir() {
			http.ServeFileFS(writer, request, frontendFiles, "index.html")
			return
		}

		fileServer.ServeHTTP(writer, request)
	})
}
