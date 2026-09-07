package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"io"
	"context"
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
	gameContext, gameCancel := context.WithCancel(context.Background())

	// Kick off game loop in a separate goroutine
	go gameState.Run(gameContext)

	// Set server endpoint handlers
	apiState := api.InitState(gameState)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth", apiState.HandlePostAuth)
	mux.HandleFunc("GET /api/websocket", apiState.HandleGetWebSocket)

	// Begin server
	log.Printf("Beginning server on port %d...", env.Port)
	address := fmt.Sprintf(":%d", env.Port)
	serveErr := http.ListenAndServe(address, corsMiddleware(mux))
	if serveErr != nil && serveErr != http.ErrServerClosed {
		log.Printf("Server failed with error: %s", serveErr.Error())
	}

	// Trigger the game loop goroutine to end
	gameCancel()
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
