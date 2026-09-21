package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"
	"net/http"
	"mud/api"
	"mud/game"
)

const MUD_LOG_FOLDER = "./logs"

func main() {
	// Init logger
	logfile := initLogger()
	defer logfile.Close()

	// Load env
	api.LoadEnv()
	env := api.GetEnv()

	// Init gameState
	gameContext, gameCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer gameCancel()

	// Set server endpoint handlers
	gameState := game.GameStateInit()
	apiState := api.InitState(gameState)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", apiState.HandleAuthLogin)
	mux.HandleFunc("/api/auth/callback", apiState.HandleAuthCallback)
	if env.ENABLE_DEBUG_AUTH {
		mux.HandleFunc("/api/auth/debug", apiState.HandleDebugLogin)
	}
	mux.HandleFunc("/api/websocket", apiState.HandleGetWebSocket)

	// Kick off server in a separate goroutine
	// Begin server
	log.Printf("Beginning server on port %d...", env.PORT)
	server := &http.Server {
		Addr: fmt.Sprintf(":%d", env.PORT),
		Handler: corsMiddleware(mux),
	}
	go runHttpServer(server)

	// Kick off game loop on main thread
	gameState.Run(gameContext)

	// Tell the HTTP server to shutdown gracefully
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err := server.Shutdown(shutdownContext)
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
	// (I don't know why that's the correct format string to use, but it is)
	timestamp := time.Now().Format("2006-01-02T15:04:05")
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
