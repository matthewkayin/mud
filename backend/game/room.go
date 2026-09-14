package game

import (
	"log"
	"sort"
)

const ROOM_NONE int = -1

type Room struct {
	Name string
	Description string

	ExitNorth int
	ExitSouth int
	ExitEast int
	ExitWest int

	occupants []MobHandle
	Inventory ItemList
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.occupants = append(room.occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := -1
	for index, occupant := range room.occupants {
		if occupant.Equals(handle) {
			occupantIndex = index
			break
		}
	}
	if occupantIndex == -1 {
		log.Printf("Warning: Tried to remove occupant with handle %d:%d from room %s, but no such occupant was found.",
			handle.id, handle.generation, room.Name)
	}

	room.RemoveOccupantByIndex(occupantIndex)
}

func (room *Room) RemoveOccupantByIndex(index int) {
	lastIndex := len(room.occupants) - 1
	room.occupants[index] = room.occupants[lastIndex]
	room.occupants = room.occupants[:lastIndex]
}

func (room *Room) Update(gameState *GameState) {
	// Sort combatants by initiative order
	sort.Slice(room.occupants, func(i int, j int) bool {
		mobI := gameState.world.Mobs.Get(room.occupants[i])
		mobJ := gameState.world.Mobs.Get(room.occupants[j])
		return mobI.Data.Agility() > mobJ.Data.Agility()
	})

	// Occupant update and combat
	occupantIndex := 0
	for occupantIndex < len(room.occupants) {
		// Get occupant mob
		occupantHandle := room.occupants[occupantIndex]
		occupantMob := gameState.world.Mobs.Get(occupantHandle)

		// Check for mob death
		if occupantMob.IsDead() {
			room.RemoveOccupantByIndex(occupantIndex)
			if occupantMob.player != nil {
				occupantMob.player.onDeath(gameState)
			}
			gameState.world.Mobs.Remove(occupantHandle)
			continue
		}

		// Update mob
		occupantMob.Update(gameState)
		occupantIndex += 1
	}
}

func (room *Room) broadcast(gameState *GameState, message string) {
	for _, occupantHandle := range room.occupants {
		occupantMob := gameState.world.Mobs.Get(occupantHandle)
		if occupantMob.player == nil {
			continue
		}

		*occupantMob.player.inbox <- message
	}
}
