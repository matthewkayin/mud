package game

type EventType int
const (
	EVENT_TYPE_MOB_MOVE = iota
	EVENT_TYPE_MOB_DEATH
	EVENT_TYPE_PLAYER_LOGOUT
	EVENT_TYPE_MOB_SET_TARGET
	EVENT_TYPE_COUNT
)

type Event struct {
	eventType EventType
	data any
}

type EventMobMove struct {
	mobHandle MobHandle
	fromRoom int
	toRoom int
}

type EventMobDeath struct {
	mobHandle MobHandle
}

type EventPlayerLogout struct {
	playerId int
}

type EventMobSetTarget struct {
	attacker MobHandle
	defender MobHandle
}

func (gameState *GameState) addEventListener(eventType EventType, listener func (gameState *GameState, event Event)) {
	gameState.eventListeners[eventType] = append(gameState.eventListeners[eventType], listener)
}

func (gameState *GameState) fireEvent(event Event) {
	for _, listener := range gameState.eventListeners[event.eventType] {
		listener(gameState, event)
	}
}
