package world

import (
	"fmt"
	"log"
	"sort"
	"slices"

	"mud/util"
	"github.com/mmcdole/lunar"
)

const ROOM_NONE int = -1
const ROOM_DATA_FOLDER = "rooms"
const ROOM_RESET_FREQUENCY uint32 = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type RoomPosition struct {
	X int
	Y int
	Z int
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
	npcSpawners []NpcSpawner

	State lua.Value
	Occupants []MobHandle
	Chests []Chest
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.Occupants = append(room.Occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := slices.IndexFunc(room.Occupants, func (occupantHandle MobHandle) bool {
		return handle == occupantHandle
	})
	if occupantIndex == -1 {
		log.Printf("warn: tried to remove occupant with handle %s from %s, but no such occupant was found.",
			handle.String(), room.Name)
		return
	}

	room.removeOccupantByIndex(occupantIndex)
}

func (room *Room) removeOccupantByIndex(index int) {
	room.Occupants = util.RemoveAtUnordered(room.Occupants, index)
}

func (room *Room) Reset(world *World) {
	room.lootTable.replenish(world, &room.Inventory, 0)

	if room.reset != nil {
		world.ScriptCall(room.reset.Value(), "reset")
	}
}

func (room *Room) update(world *World) {
	// Update NPC spawners
	for index := range len(room.npcSpawners) {
		room.npcSpawners[index].update(world, room.index)
	}

	// Update occupants
	occupants := room.sortOccupantsByInitiativeOrder(world)
	for _, occupantHandle := range occupants {
		occupant := world.Mobs.Get(occupantHandle)
		if occupant.isDead() {
			continue
		}

		occupant.update(world)
	}

	// Remove dead occupants
	index := 0
	for index < len(room.Occupants) {
		occupantHandle := room.Occupants[index]
		occupant := world.Mobs.Get(occupantHandle)
		if occupant.isDead() {
			room.Occupants = slices.Delete(room.Occupants, index, index + 1)
			room.grantExperienceFrom(world, occupant)
			room.createCorpse(world, occupant)
			world.Mobs.Remove(occupantHandle)
		} else {
			index++
		}
	}

	// Update chests
	index = 0
	for index < len(room.Chests) {
		chest := &room.Chests[index]
		chest.update()

		if chest.isRottedCorpse() {
			room.Chests = util.RemoveAtUnordered(room.Chests, index)
		} else {
			index++
		}
	}
}

func (room *Room) sortOccupantsByInitiativeOrder(world *World) []MobHandle {
	occupants := slices.Clone(room.Occupants)
	sort.SliceStable(occupants, func(i int, j int) bool {
		mobI := world.Mobs.Get(occupants[i])
		mobJ := world.Mobs.Get(occupants[j])

		// The sort is stable, so mobs with the same initiative keep their order in room.Occupants.
		// This ensures a fair initiative over subsequent rounds, otherwise we end up with
		// unfair initiatives like A B B A, where B gets two turns in a row over A
		if mobI.Stat(STAT_INITIATIVE) == mobJ.Stat(STAT_INITIATIVE) {
			return false
		}

		return mobI.Stat(STAT_INITIATIVE) > mobJ.Stat(STAT_INITIATIVE)
	})

	return occupants
}

func (room *Room) createCorpse(world *World, mob *Mob) {
	var corpse Chest
	if mob.IsPlayer() {
		mob.Character.dumpEquipment(world)
		corpse = NewPlayerCorpse(mob.Character)
	} else {
		corpse = NewCorpse(mob.Name, &mob.Npc.data.loot)
	}
	room.Chests = append(room.Chests, corpse)
}

func (room *Room) grantExperienceFrom(world *World, deadMob *Mob) {
	// Don't grant experience for non-NPC kills
	if !deadMob.IsNpc() {
		return
	}

	// Gather a list of all player mobs
	players := make([]*Mob, 0, len(room.Occupants))
	for index := range room.Occupants {
		occupant := world.Mobs.Get(room.Occupants[index])
		if occupant.IsPlayer() {
			players = append(players, occupant)
		}
	}

	expPerPlayer := deadMob.Npc.data.experienceWorth / int32(len(players))
	for _, player := range players {
		player.Character.grantExperience(world, expPerPlayer)
	}
}

func (position RoomPosition) String() string {
	return fmt.Sprintf("<%d, %d, %d>", position.X, position.Y, position.Z)
}
