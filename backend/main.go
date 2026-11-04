package main

import (
	"os"
	"io"
	"fmt"
	"log"
	"time"
	"context"
	"syscall"
	"net/http"
	"os/signal"
	"mud/api"
	"mud/env"
	"mud/game"
)

const MUD_LOG_FOLDER = "logs"

// Set at build time with -ldflags "-X main.version=..."
var version = "dev"

func main() {
	// Init logger
	logfile := initLogger()
	defer logfile.Close()

	// Load env
	env.Load()
	env := env.Get()
	log.Printf("Starting %s version %s...", env.NAME, version)

	// Create save folders
	err := os.MkdirAll(api.AccountFolderPath(), 0700)
	if err != nil {
		log.Fatalf("Failed to create account save folder: %s", err.Error())
	}

	// Init gamestate
	gameContext, gameCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer gameCancel()
	gamestate := game.Init(version)

	// Set server endpoint handlers
	apiState := api.InitState(gamestate)
	mux := http.NewServeMux()
	apiState.RegisterAuthRoutes(mux)
	// Unknown API paths shouldn't fall through to the frontend once it is served from here
	mux.Handle("/api/", http.NotFoundHandler())

	// Begin server
	log.Printf("Beginning server on port %d...", env.PORT)
	server := &http.Server {
		Addr: fmt.Sprintf(":%d", env.PORT),
		Handler: mux,
	}
	go runHttpServer(server)

	// Kick off game loop on main thread
	gamestate.Run(gameContext)

	// Tell the HTTP server to shutdown gracefully
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancelShutdown()

	err = server.Shutdown(shutdownContext)
	if err != nil {
		log.Fatalf("HTTP server did not shutdown gracefully: %s", err.Error())
	}

	log.Printf("Server shutdown gracefully.")
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

func runHttpServer(server *http.Server) {
	serveErr := server.ListenAndServe()
	if serveErr != nil && serveErr != http.ErrServerClosed {
		log.Printf("Server failed with error: %s", serveErr.Error())
	}
}
