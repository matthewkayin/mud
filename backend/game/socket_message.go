package game

// The player's current stats, sent with heartbeats while they are in-game
type PlayerStatus struct {
	Health int32 `json:"health"`
	MaxHealth int32 `json:"maxHealth"`
	Mana int32 `json:"mana"`
	MaxMana int32 `json:"maxMana"`
	PendingAction bool `json:"action"`
}

// A message sent to the client. Message is nil for a heartbeat.
// The PlayerStatus fields are left out of the JSON when PlayerStatus is nil.
type SocketMessage struct {
	Message *string `json:"message"`
	*PlayerStatus
}
