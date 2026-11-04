package game

type SocketEventType int
const (
	SOCKET_EVENT_TYPE_CONNECT SocketEventType = iota
	SOCKET_EVENT_TYPE_DISCONNECT
	SOCKET_EVENT_TYPE_COMMAND
)

type SocketEvent struct {
	Type SocketEventType
	PlayerKey string
	// Used to identify the connection on CONNECT and DISCONNECT
	Inbox *chan SocketMessage
	// Only used by COMMAND events
	Command string
}

// A message sent to the client
type SocketMessage struct {
	Message *string `json:"message"`
}
