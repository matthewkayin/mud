package world

import (
	"log"
	"slices"

	"mud/util"
	"github.com/mmcdole/lunar"
)

const ROOM_NONE int = -1
const ROOM_DATA_FOLDER = "rooms"
const ROOM_RESET_FREQUENCY uint32 = (5 * 60) / WORLD_SECONDS_PER_UPDATE

const ROOM_NPC_RESPAWN_TIMER_NOT_STARTED int = -1
const ROOM_NPC_RESPAWN_DURATION int = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type RoomPosition struct {
	X int
	Y int
	Z int
}

type RoomNpcSpawner struct {
	npcData *NpcData
	mobHandle MobHandle
	respawnDuration int
	respawnTimer int
}

type Room struct {
	index int
	Name string
	Description string
	Position RoomPosition

	Details map[string]string
	reset *lua.Function
	Interactions map[string]*lua.Function
	Exits [DIRECTION_COUNT]int

	Inventory Inventory
	lootTable LootTable
	npcSpawners []RoomNpcSpawner

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
	room.Occupants = util.RemoveAtUnordered(room.Occupants, index)
}

func (room *Room) Reset(world *World) {
	room.lootTable.replenish(world, &room.Inventory)

	if room.reset != nil {
		world.ScriptCall(room.reset.Value(), "reset")
	}
}

func (room *Room) update(world *World) {
	// Update NPC spawners
	for index := range len(room.npcSpawners) {
		spawner := &room.npcSpawners[index]

		// If the timer has not started, it means we should have a mob
		if spawner.respawnTimer == ROOM_NPC_RESPAWN_TIMER_NOT_STARTED {
			// Check to see if our mob is still alive
			_, exists := world.Mobs.GetIfExists(spawner.mobHandle)
			if !exists {
				spawner.respawnTimer = spawner.respawnDuration
			}
		}

		// If the timer has started, count it down
		if spawner.respawnDuration != ROOM_NPC_RESPAWN_TIMER_NOT_STARTED {
			spawner.respawnTimer--
			if spawner.respawnTimer <= 0 {
				spawner.mobHandle = world.spawnNpc(NpcSpawnParams {
					data: spawner.npcData,
					room: room.index,
					movementType: NPC_MOVEMENT_TYPE_STATIONARY,
				})
				spawner.respawnTimer = ROOM_NPC_RESPAWN_TIMER_NOT_STARTED
			}
		}
	}

	// Update occupants
	for index := range len(room.Occupants) {
		occupant := world.Mobs.Get(room.Occupants[index])
		if occupant.isDead() {
			continue
		}

		occupant.update(world)
	}

	// Remove dead occupants
	index := 0
	for index < len(room.Occupants) {
		occupant := world.Mobs.Get(room.Occupants[index])
		if occupant.isDead() {
			util.RemoveAtUnordered(room.Occupants, index)
		} else {
			index++
		}
	}
}
