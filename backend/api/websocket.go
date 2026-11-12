package api

import (
	"context"
	"log"
	"mud/game"
	"net/http"
	"regexp"
	"strings"
	"time"
	"encoding/json"

	"mud/env"

	"github.com/coder/websocket"
)

const SOCKET_WRITE_TIMEOUT = 5 * time.Second
const SOCKET_PING_INTERVAL = 20 * time.Second
const SOCKET_PING_TIMEOUT = 10 * time.Second

var SOCKET_COMMAND_FORBIDDEN_CHARS = regexp.MustCompile(`[^\x20-\x7E]|[<>\[\]{}\\]`)

func (apiState *ApiState) HandleGetWebSocket(writer http.ResponseWriter, request *http.Request) {
	playerKey, authenticated := apiState.getWebSocketPlayerKey(request)
	if !authenticated {
		http.Error(writer, "User unauthorized.", http.StatusUnauthorized)
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	connection, err := websocket.Accept(writer, request, &websocket.AcceptOptions {
		// Same-origin requests are always accepted. These allow the Vite dev server, whose proxy rewrites the Host header.
		OriginPatterns: []string {
			"localhost:5173",
			"192.168.*:5173",
			"10.100.*:5173",
		},
	})
	if err != nil {
		// Note: we can't return HTTP response here because the connection
		// has already been upgraded to web socket even though there's an error
		log.Printf("Error accepting websocket: %s", err.Error())
		return
	}

	log.Printf("Player %s connected.", playerKey)

	// Create the player's inbox. It will be owned and closed by the gamestate
	inbox := make(chan game.SocketMessage, 100)
	go runSocketWriteLoop(connection, inbox, playerKey)

	// Send connect event
	apiState.gamestate.SocketEvents <- game.SocketEvent {
		Type: game.SOCKET_EVENT_TYPE_CONNECT,
		PlayerKey: playerKey,
		Inbox: &inbox,
	}

	// Start the ping loop
	// The ping only receives its pong when the read loop is running, so the ping starts first
	// The ping loop stops when the read loop exits
	pingContext, cancelPing := context.WithCancel(request.Context())
	defer cancelPing()
	go runSocketPingLoop(pingContext, connection, playerKey)

	// Run the read loop until the connection closes
	readloop:
	for {
		messageType, messageData, err := connection.Read(request.Context())
		if err != nil {
			log.Printf("Player %s disconnected.", playerKey)
			break readloop
		}

		if messageType == websocket.MessageText {
			apiState.gamestate.SocketEvents <- game.SocketEvent {
				Type: game.SOCKET_EVENT_TYPE_COMMAND,
				PlayerKey: playerKey,
				Command: sanitizeCommand(string(messageData)),
			}
		}
	}
}

func runSocketWriteLoop(connection *websocket.Conn, inbox chan game.SocketMessage, playerKey string) {
	connectionOpen := true

	for message := range inbox {
		if !connectionOpen {
			continue
		}

		messageJson, err := json.Marshal(message)
		if err != nil {
			log.Printf("Failed to encode message for player %s: %s", playerKey, err.Error())
			continue
		}

		writeContext, cancel := context.WithTimeout(context.Background(), SOCKET_WRITE_TIMEOUT)
		err = connection.Write(writeContext, websocket.MessageText, messageJson)
		cancel()
		if err != nil {
			log.Printf("Failed to write to player %s: %s", playerKey, err.Error())

			// Close the connection
			// This ends the read loop and disconnects the player
			connection.CloseNow()
			connectionOpen = false
		}
	}

	// Inbox was closed by the game loop, so the session is over
	if connectionOpen {
		connection.Close(websocket.StatusNormalClosure, "session ended")
	}
}

func runSocketPingLoop(ctx context.Context, connection *websocket.Conn, playerKey string) {
	ticker := time.NewTicker(SOCKET_PING_INTERVAL)
	defer ticker.Stop()

	for {
		select {
			case <- ctx.Done():
				return

			case <- ticker.C:
				pingContext, cancel := context.WithTimeout(ctx, SOCKET_PING_TIMEOUT)
				err := connection.Ping(pingContext)
				cancel()
				if err != nil {
					// The ping loop was stopped because the connection is already closed
					if ctx.Err() != nil {
						return
					}

					log.Printf("Ping to player %s failed: %s", playerKey, err.Error())
					connection.CloseNow()
					return
				}
		}
	}
}

func sanitizeCommand(command string) string {
	command = strings.TrimSpace(command)
	command = SOCKET_COMMAND_FORBIDDEN_CHARS.ReplaceAllString(command, "")

	return command
}

// With debug auth enabled, a debugUser query param names the player directly, with no session.
// Otherwise the player comes from the session cookie.
func (apiState *ApiState) getWebSocketPlayerKey(request *http.Request) (string, bool) {
	debugUsername := request.URL.Query().Get("debugUser")
	if !env.Get().ENABLE_DEBUG_AUTH || debugUsername == "" {
		return apiState.getSessionKey(request)
	}

	// The player key is used in save file paths, so it must be a valid username
	err := validateUsername(debugUsername)
	if err != nil {
		return "", false
	}

	return strings.ToLower(debugUsername), true
}
