package world

import (
	"fmt"
	"log"
	"errors"
)

const ROOM_NONE int = -1

const CHEST_DOES_NOT_DECAY = -1
const CHEST_CORPOSE_DECAY_DURATION = 30 / WORLD_SECONDS_PER_UPDATE

type Chest struct {
	Name string
	DecayTimer int
	Inventory Inventory
}

type Room struct {
	Name string
	Description string
	Exits [DIRECTION_COUNT]int
	ExitIsLocked [DIRECTION_COUNT]bool
	IsSafeZone bool
	Inventory Inventory
	Chests []Chest

	Occupants []MobHandle
}

func (room *Room) MoveOccupant(world *World, occupantHandle MobHandle, direction Direction) error {
	// Check if there is an exit in that direction
	newRoomIndex := room.Exits[direction]
	if newRoomIndex == ROOM_NONE {
		return errors.New("There is no exit in that direction.")
	}

	// Check if the exit is locked
	if room.ExitIsLocked[direction] {
		return fmt.Errorf("The %s exit is locked.", DirectionToString(direction))
	}

	// Get a pointer to the new room
	newRoom := &world.Rooms[newRoomIndex]
	occupantMob := world.Mobs.Get(occupantHandle)
	oldRoomIndex := occupantMob.Data.Room

	// Move the occupant

	room.RemoveOccupant(occupantHandle)

	// Note that the order matters here, we don't want to send these messages to the moving
	// player. Since the occupantHandle is in neither room at this point, the broadcast
	// function will not send the messages into the occupant's inbox
	world.messageRoom(oldRoomIndex, fmt.Sprintf("%s left the room.", occupantMob.Data.Name))
	world.messageRoom(newRoomIndex, fmt.Sprintf("%s entered the room.", occupantMob.Data.Name))

	newRoom.AddOccupant(occupantHandle)
	occupantMob.Data.Room = newRoomIndex

	// Fire event
	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_MOVE,
		Data: EventMobMove {
			MobHandle: occupantHandle,
			FromRoom: oldRoomIndex,
			ToRoom: newRoomIndex,
		},
	})

	return nil
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.Occupants = append(room.Occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := -1
	for index, occupant := range room.Occupants {
		if occupant == handle {
			occupantIndex = index
			break
		}
	}
	if occupantIndex == -1 {
		log.Printf("Warning: Tried to remove occupant with handle %d:%d from room %s, but no such occupant was found.",
			handle.Id, handle.Generation, room.Name)
	}

	room.RemoveOccupantByIndex(occupantIndex)
}

func (room *Room) RemoveOccupantByIndex(index int) {
	lastIndex := len(room.Occupants) - 1
	room.Occupants[index] = room.Occupants[lastIndex]
	room.Occupants = room.Occupants[:lastIndex]
}
