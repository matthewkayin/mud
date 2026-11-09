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

type SocketPlayerStatus struct {
	Health int32 `json:"hp"`
	MaxHealth int32 `json:"maxHp"`
	Mana int32 `json:"mp"`
	MaxMana int32 `json:"maxMp"`
	PendingAction bool `json:"action"`
}

// A message sent to the client
type SocketMessage struct {
	Message *string `json:"message"`
	Status *SocketPlayerStatus `json:"status,omitempty"`
	Inventory []string `json:"inv,omitempty"`
	Map []int `json:"map,omitempty"`
}
