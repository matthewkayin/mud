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

type SocketInventoryItem struct {
	Name string `json:"name"`
	Amount int32 `json:"amount"`
	Size int32 `json:"size"`
}

// The player's inventory, which the client formats into a table for the view panel
type SocketInventory struct {
	Size int32 `json:"size"`
	Capacity int32 `json:"capacity"`
	Items []SocketInventoryItem `json:"items"`
}

// One equipment slot, which the client formats into a table for the view panel
// Name is empty when nothing is equipped in the slot
// When a two-handed weapon is equipped, the main hand's Slot is "Both Hands" and the off hand's Slot and Name are empty
type SocketEquipmentSlot struct {
	Slot string `json:"slot"`
	Name string `json:"name"`
}

// One mob in the player's room, for the client's mobs table
// Name and Target include the mob's fuzzy number when several mobs in the room share its name (see getRoomMobLabels)
// Target is empty when the mob isn't targeting a mob in the room
type SocketRoomMob struct {
	Name string `json:"name"`
	Condition string `json:"condition"`
	Target string `json:"target"`
}

// A named number on the character sheet, such as an attribute or stat
// IsPercentage is true when the client should show the value with a "%" after it
type SocketSheetValue struct {
	Name string `json:"name"`
	Value int32 `json:"value"`
	IsPercentage bool `json:"percent"`
}

// The player's character sheet, which the client formats for the view panel's stats mode
// Experience is the experience earned toward the next level, out of ExperienceToNextLevel
type SocketCharacterSheet struct {
	Name string `json:"name"`
	Race string `json:"race"`
	Class string `json:"class"`
	Level int32 `json:"level"`
	Experience int32 `json:"exp"`
	ExperienceToNextLevel int32 `json:"expNext"`
	Attributes []SocketSheetValue `json:"attributes"`
	Stats []SocketSheetValue `json:"stats"`
	Skills []string `json:"skills"`
}

// The player's spells, which the client formats into tables for the view panel's spells mode
// Prepared and Known are never nil, so empty lists are sent as [] rather than null
// Slots is the number of spells the character can have prepared at once
type SocketSpells struct {
	Prepared []string `json:"prepared"`
	Known []string `json:"known"`
	Slots int32 `json:"slots"`
}

// A message sent to the client
type SocketMessage struct {
	Message *string `json:"message"`
	Status *SocketPlayerStatus `json:"status,omitempty"`
	Inventory *SocketInventory `json:"inv,omitempty"`
	Equipment []SocketEquipmentSlot `json:"equip,omitempty"`
	Map []int `json:"map,omitempty"`
	RoomMobs []SocketRoomMob `json:"mobs,omitempty"`
	Sheet *SocketCharacterSheet `json:"sheet,omitempty"`
	Spells *SocketSpells `json:"spells,omitempty"`
}
