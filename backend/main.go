package main

import (
	"fmt"
	"log"
	"net/http"
	"mud/mudenv"
	"mud/api"
)

type ErrorResponse struct {
	error string
}

func main() {
	log.SetFlags(0)

	mudenv.LoadFromFile("env.json")
	env := mudenv.Get()

	log.Printf("Beginning server on port %d...", env.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth", api.HandlePostAuth)
	mux.HandleFunc("GET /api/websocket", api.HandleGetWebSocket)

	address := fmt.Sprintf(":%d", env.Port)
	serveErr := http.ListenAndServe(address, corsMiddleware(mux))
	if serveErr != nil && serveErr != http.ErrServerClosed {
		log.Fatal(serveErr.Error())
	}
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
