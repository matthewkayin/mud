package api

import (
	"context"
	"log"
	"strings"
	"net/http"
	"mud/game"
	"github.com/coder/websocket"
)

func (apiState* ApiState) HandleGetWebSocket(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked /api/websocket")

	// Get the auth token from the query param
	sessionToken, err := request.Cookie(MUD_SESSION_COOKIE_NAME)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return
	}

	// Check the list of active auth tokens
	// TODO: handle token expiration and clear the cache?
	apiState.tokenToIdMutex.RLock()
	userId, authenticated := apiState.tokenToIdMap[sessionToken.Value]
	apiState.tokenToIdMutex.RUnlock()

	if !authenticated {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("Authenticated user %d", userId)

	// Upgrade the HTTP connection to a WebSocket connection
	connection, acceptError := websocket.Accept(writer, request, &websocket.AcceptOptions {
		// TODO: configure this for prod
		OriginPatterns: []string {
			"localhost:5173",
			"192.168.*:5173",
			"10.100.*:5173",
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

	apiState.gameState.RemovePlayer(userId)
}
