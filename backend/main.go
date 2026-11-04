package main

import (
	"os"
	"io"
	"fmt"
	"log"
	"time"
	"strings"
	"net/http"
	"mud/env"
)

const MUD_LOG_FOLDER = "logs"

// Set at build time with -ldflags "-X main.version=..."
var version = "dev"

func main() {
	// Init logger
	logfile := initLogger()
	defer logfile.Close()

	env.Load()
	env := env.Get()
	log.Printf("Starting %s version %s...", env.NAME, version)
}

func initLogger() *os.File {
	// Create log folder
	err := os.MkdirAll(MUD_LOG_FOLDER, 0755)
	if err != nil {
		log.Fatalf("Failed to create log folder: %s", err.Error())
	}

	// Determine logfile path
	timestamp := time.Now().Format("2006-01-02T150405")
	logfilePath := fmt.Sprintf("%s/%s.log", MUD_LOG_FOLDER, timestamp)

	// Open logfile
	fileOpenFlags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	logfile, err := os.OpenFile(logfilePath, fileOpenFlags, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err.Error())
	}

	// Set logger to write both to stdout and file
	multiwriter := io.MultiWriter(os.Stdout, logfile)
	log.SetOutput(multiwriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logfile
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func (writer http.ResponseWriter, request *http.Request) {
		// The editor's routes write files, so they don't allow other origins
		if strings.HasPrefix(request.URL.Path, "/api/editor/") {
			next.ServeHTTP(writer, request)
			return
		}

		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		writer.Header().Set("Access-Control-Max-Age", "3600")

		if request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func runHttpServer(server *http.Server) {
	serveErr := server.ListenAndServe()
	if serveErr != nil && serveErr != http.ErrServerClosed {
		log.Printf("Server failed with error: %s", serveErr.Error())
	}
}
