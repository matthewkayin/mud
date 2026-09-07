package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"mud/core"
	"mud/game"
	"net/http"

	"github.com/coder/websocket"
)

type recurseApiGetProfilesResponseSuccess struct {
	Id int `json:"id"`
}

func (apiState* ApiState) HandleGetWebSocket(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked GET /api/websocket")
	env := core.GetEnv()

	// Get the auth token from the query param
	token := request.URL.Query().Get("token")

	// Check the list of active auth tokens
	apiState.tokenToIdMutex.RLock()
	userId, authenticated := apiState.tokenToIdMap[token]
	apiState.tokenToIdMutex.RUnlock()

	// TODO: handle token expiration and clear the cache?

	// If it's not, then try querying the RC API for this user
	if !authenticated || token == "" {
		log.Printf("User ID not found in cache. Querying RC API...")

		// Build a GET request to the RC API
		recurseRequest, newRequestError := http.NewRequest("GET", fmt.Sprintf("%s/profiles/me", env.RecurseApiUrl), nil)
		if newRequestError != nil {
			http.Error(writer, newRequestError.Error(), http.StatusInternalServerError)
			return
		}
		recurseRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Make the GET request
		httpClient := &http.Client{}
		recurseResponse, recurseError := httpClient.Do(recurseRequest)
		if recurseError != nil {
			http.Error(writer, recurseError.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			io.Copy(io.Discard, recurseResponse.Body)
			recurseResponse.Body.Close()
		}()

		// For some reason RC returns 404 when you don't provide an auth token
		if recurseResponse.StatusCode == http.StatusNotFound || recurseResponse.StatusCode == http.StatusUnauthorized {
			http.Error(writer, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		// Get the contents of the response body
		recurseResponseBodyBytes, err := io.ReadAll(recurseResponse.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert the response from JSON
		var recurseResponseBody recurseApiGetProfilesResponseSuccess
		unmarshalError := json.Unmarshal(recurseResponseBodyBytes, &recurseResponseBody)
		if unmarshalError != nil {
			http.Error(writer, unmarshalError.Error(), http.StatusInternalServerError)
			return
		}

		// Cache the ID for later
		apiState.tokenToIdMap[token] = recurseResponseBody.Id
		userId = recurseResponseBody.Id
	}

	log.Printf("Authenticated user %d", userId)

	// Upgrade the HTTP connection to a WebSocket connection
	connection, acceptError := websocket.Accept(writer, request, &websocket.AcceptOptions {
		OriginPatterns: []string {
			"localhost:5173",
		},
	})
	if acceptError != nil {
		// Note: we can't return HTTP response here because the connection
		// has already been upgraded to web socket even though there's an error
		log.Printf("Accept error: %s", acceptError.Error())
		return
	}

	log.Printf("Player %d connected.", userId)

	// Kick off write loop in a separate goroutine
	writerContext, cancelWriter := context.WithCancel(request.Context())
	defer cancelWriter()
	go apiState.runSocketWriteLoop(writerContext, connection, userId)

	// Run read loop until connection closes
	readLoop:
	for {
		messageType, messageData, err := connection.Read(request.Context())
		if err != nil {
			log.Printf("Player %d disconnected.", userId)
			break readLoop
		}

		if messageType == websocket.MessageText {
			apiState.gameState.Commands <- game.Command {
				PlayerId: userId,
				Payload: strings.TrimSpace(string(messageData)),
			}
		}
	}
}

func (apiState *ApiState) runSocketWriteLoop(ctx context.Context, connection *websocket.Conn, userId int) {
	inbox := make(chan string, 100)
	apiState.gameState.RegisterPlayer(userId, &inbox)

	writeLoop:
	for {
		select {
			case <- ctx.Done():
				break writeLoop
			case message, ok := <- inbox:
				if !ok {
					break writeLoop
				}

				err := connection.Write(ctx, websocket.MessageText, []byte(message))
				if err != nil {
					log.Printf("Failed to write to player %d: %s", userId, err.Error())
					break writeLoop
				}
		}
	}

	// TODO: on exit, unregister the player
}
