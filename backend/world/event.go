package world

import "slices"

type EventType int
const (
	EVENT_TYPE_MESSAGE = iota
	EVENT_TYPE_MOB_MOVE
	EVENT_TYPE_MOB_DEATH
	EVENT_TYPE_MOB_SET_TARGET
	EVENT_TYPE_COUNT
)

type Event struct {
	EventType EventType
	Data any
}

type EventMessage struct {
	ToPlayers []int
	Message string
}

type EventMobMove struct {
	MobHandle MobHandle
	FromRoom int
	ToRoom int
}

type EventMobDeath struct {
	PlayerId int
	MobHandle MobHandle
}

type EventMobSetTarget struct {
	Attacker MobHandle
	Defender MobHandle
}

type MessageRoomOptions struct {
	roomIndex int
	ignore []MobHandle
	message string
}

func (world *World) pushEvent(event Event) {
	world.Events = append(world.Events, event)
}

// Adds a message event to the world event queue to a specific player
// Messages sent using this function will be processed immediately after world update
// This means you should only use this during world update or the message timing will be delayed
func (world *World) messagePlayer(playerId int, message string) {
	world.Events = append(world.Events, Event {
		EventType: EVENT_TYPE_MESSAGE,
		Data: EventMessage {
			ToPlayers: []int { playerId },
			Message: message,
		},
	})
}

// Adds a message event to the world event queue to all players in the room
// Messages sent using this function will be processed immediately after world update
// This means you should only use this during world update or the message timing will be delayed
func (world *World) messageRoom(roomIndex int, message string) {
	world.messageRoomWithOptions(MessageRoomOptions {
		roomIndex: roomIndex,
		ignore: []MobHandle {},
		message: message,
	})
}

func (world *World) messageRoomWithOptions(options MessageRoomOptions) {
	room := &world.Rooms[options.roomIndex]

	toPlayers := make([]int, 0, len(room.Occupants))
	for _, mobHandle := range room.Occupants {
		// Don't send messages to mobs on the ignore list
		if slices.Contains(options.ignore, mobHandle) {
			continue
		}

		// Don't send messages to NPCs
		mob := world.Mobs.Get(mobHandle)
		if mob.PlayerCharacter == nil {
			continue
		}

		toPlayers = append(toPlayers, mob.PlayerCharacter.PlayerId)
	}

	if len(toPlayers) == 0 {
		return
	}

	world.Events = append(world.Events, Event {
		EventType: EVENT_TYPE_MESSAGE,
		Data: EventMessage {
			ToPlayers: toPlayers,
			Message: options.message,
		},
	})
}
