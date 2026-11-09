package world

import (
	"log"
	"slices"

	"github.com/mmcdole/lunar"
)

const ROOM_NONE int = -1
const ROOM_DATA_FOLDER = "rooms"
const ROOM_RESET_FREQUENCY = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type RoomPosition struct {
	X int
	Y int
	Z int
}

type Room struct {
	Name string
	Description string
	Position RoomPosition

	Details map[string]string
	reset *lua.Function
	Interactions map[string]*lua.Function
	Exits [DIRECTION_COUNT]int

	Inventory Inventory
	lootTable LootTable

	State lua.Value
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
		return
	}

	room.removeOccupantByIndex(occupantIndex)
}

func (room *Room) removeOccupantByIndex(index int) {
	lastIndex := len(room.Occupants) - 1
	room.Occupants[index] = room.Occupants[lastIndex]
	room.Occupants = room.Occupants[:lastIndex]
}

func (room *Room) Reset(world *World) {
	room.lootTable.replenish(world, &room.Inventory)

	if room.reset != nil {
		world.ScriptCall(room.reset.Value(), "reset")
	}
}
