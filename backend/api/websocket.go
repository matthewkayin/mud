package api

import (
	"context"
	"log"
	"strings"
	"time"
	"net/http"
	"mud/game"
	"github.com/coder/websocket"
)

const SOCKET_WRITE_TIMEOUT = 5 * time.Second

func (apiState* ApiState) HandleGetWebSocket(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked /api/websocket")

	userId, authenticated := apiState.getAuthenticatedUserId(request)
	if !authenticated {
		http.Error(writer, "User unauthorized.", http.StatusUnauthorized)
		return
	}

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

	// The game loop owns the inbox and closes it when the player's session ends.
	// The write loop listens to the inbox it until the inbox is closed by the game loop.
	inbox := make(chan string, 100)
	go runSocketWriteLoop(connection, inbox, userId)

	apiState.gamestate.SocketEvents <- game.SocketEvent {
		Type: game.SOCKET_EVENT_TYPE_CONNECT,
		PlayerId: userId,
		Inbox: &inbox,
	}

	// Run read loop until connection closes
	readLoop:
	for {
		messageType, messageData, err := connection.Read(request.Context())
		if err != nil {
			log.Printf("Player %d disconnected.", userId)
			break readLoop
		}

		if messageType == websocket.MessageText {
			apiState.gamestate.SocketEvents <- game.SocketEvent {
				Type: game.SOCKET_EVENT_TYPE_COMMAND,
				PlayerId: userId,
				Command: strings.TrimSpace(string(messageData)),
			}
		}
	}

	apiState.gamestate.SocketEvents <- game.SocketEvent {
		Type: game.SOCKET_EVENT_TYPE_DISCONNECT,
		PlayerId: userId,
		Inbox: &inbox,
	}
}

func runSocketWriteLoop(connection *websocket.Conn, inbox chan string, userId int) {
	connectionOpen := true

	// Keep draining the inbox even after the connection fails,
	// so that the game loop never blocks sending to a dead connection
	for message := range inbox {
		if !connectionOpen {
			continue
		}

		writeContext, cancel := context.WithTimeout(context.Background(), SOCKET_WRITE_TIMEOUT)
		err := connection.Write(writeContext, websocket.MessageText, []byte(message))
		cancel()
		if err != nil {
			log.Printf("Failed to write to player %d: %s", userId, err.Error())
			// Closing the connection ends the read loop, which disconnects the player
			connection.CloseNow()
			connectionOpen = false
		}
	}

	// Inbox was closed by the game loop, so the session is over
	if connectionOpen {
		connection.Close(websocket.StatusNormalClosure, "session ended")
	}
}
