package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"io"
	"context"
	"os/signal"
	"syscall"
	"net/http"
	"mud/core"
	"mud/api"
	"mud/game"
)

func main() {
	// Init logger
	logfile := initLogger()
	defer logfile.Close()

	// Load env
	core.LoadEnv("env.json")
	env := core.GetEnv()

	// Init gamestate
	gameState := game.InitState()
	gameContext, gameCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer gameCancel()

	// Set server endpoint handlers
	apiState := api.InitState(gameState)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth", apiState.HandlePostAuth)
	mux.HandleFunc("GET /api/websocket", apiState.HandleGetWebSocket)

	// Kick off server in a separate goroutine
	// Begin server
	log.Printf("Beginning server on port %d...", env.Port)
	server := &http.Server {
		Addr: fmt.Sprintf(":%d", env.Port),
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
	// Create log directory
	logFolder := "./logs"
	mkdirError := os.MkdirAll(logFolder, 0755)
	if mkdirError != nil {
		log.Fatalf("Failed to create log directory: %s", mkdirError.Error())
	}

	// Determine logfile path
	// (I don't know why that's the correct format string to use, but it is)
	timestamp := time.Now().Format("2006-01-02T15:04:05")
	logfilePath := fmt.Sprintf("%s/%s.log", logFolder, timestamp)

	// Open logfile
	fileOpenFlags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	logfile, fileOpenError := os.OpenFile(logfilePath, fileOpenFlags, 0644)
	if fileOpenError != nil {
		log.Fatalf("Failed to open log file: %s", fileOpenError.Error())
	}

	// Set logger to write to both stdout and file
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
