package world

import (
	"log"
	"slices"
)

const ROOM_NONE int = -1
const ROOM_DATA_FOLDER = "rooms"

type RoomPosition struct {
	X int
	Y int
	Z int
}

type Room struct {
	Name string
	Description string
	Position RoomPosition

	Exits [DIRECTION_COUNT]int
	Occupants []MobHandle
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.Occupants = append(room.Occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := slices.IndexFunc(room.Occupants, func (occupantHandle MobHandle) bool {
		return handle == occupantHandle
	})
	if occupantIndex == -1 {
		log.Printf("Warn: Tried to remove occupant with handle %s from %s, but no such occupant was found.",
			handle.String(), room.Name)
	}

	room.removeOccupantByIndex(occupantIndex)
}

func (room *Room) removeOccupantByIndex(index int) {
	lastIndex := len(room.Occupants) - 1
	room.Occupants[index] = room.Occupants[lastIndex]
	room.Occupants = room.Occupants[:lastIndex]
}
