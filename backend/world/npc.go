package world

import (
	"fmt"
)

// TODO: change this to a longer duration
// TODO: make this customizable per NPC?
const NPC_RESPAWN_DURATION int = 60 / WORLD_SECONDS_PER_UPDATE

type NpcBehaviorType int
const (
	NPC_BEHAVIOR_TYPE_GOBLIN = iota
)

type NpcId int
const (
	NPC_ID_GOBLIN_1 = iota
)

type Behavior interface {
	onUpdate(world *World, npc *Npc)
	onAttacked(world *World, npc *Npc)
	GetDescription(world *World, npc *Npc) (string, bool)
}

type Npc struct {
	Behavior Behavior
	Data MobData

	mobHandle MobHandle
	respawnTimer int
}

//insert an npc of a certain quantity into the world's npc array
func generateNpc(world *World, id NpcId, amount int, room int) {
	n := 0
	for n < amount {
		npc := *NPC_DATA[id]
		npc.Behavior = behaviorGoblinInit()
		npc.Data.Room = room
		world.Npcs = append(world.Npcs, npc)
		n++
	}
}

//initialize the npc in the npc array at game startup
func (npc *Npc) init(world *World) {
	npc.spawnMob(world)
}

//spawns the mob associated with the npc
func (npc *Npc) spawnMob(world *World) {
	npcMob := MobInit(&npc.Data)
	npcMob.Npc = npc
	npc.mobHandle = world.Mobs.Push(npcMob)
	npcRoom := &world.Rooms[npcMob.Data.Room]
	npcRoom.AddOccupant(npc.mobHandle)

	world.messageRoom(npcMob.Data.Room, fmt.Sprintf("%s has spawned into this room.", npcMob.Data.Name))
}

func (npc *Npc) update(world *World) {
	// If respawning, just update timer and don't do anything else
	if npc.respawnTimer != 0 {
		npc.respawnTimer--
		if npc.respawnTimer == 0 {
			npc.spawnMob(world)
		}
	}
	if npc.respawnTimer != 0 {
		return
	}

	// Check for mob death
	_, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.respawnTimer = NPC_RESPAWN_DURATION
		return
	}

	// Behavior update
	npc.Behavior.onUpdate(world, npc)
}
