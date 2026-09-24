package world

import (
	"fmt"
	"log"
)

// TODO: change this to a longer duration
// TODO: make this customizable per NPC?
const NPC_RESPAWN_DURATION int = 60 / WORLD_SECONDS_PER_UPDATE

type NpcId int
const (
	NPC_ID_GOBLIN_1 = iota
)

type NpcBehavior int
const (
	NPC_BEHAVIOR_GOBLIN = iota
)

type NpcMode int
const (
	NPC_MODE_AGGRO = iota
)


type Npc struct {
	Behavior NpcBehavior
	Mode NpcMode
	Data MobData

	mobHandle MobHandle
	respawnTimer int
}

//insert an npc of a certain quantity into the world's npc array
func generateNpc(world *World, id NpcId, amount int, room int) {
	npc := *NPC_DATA[id]
	npc.Data.Room = room
	n := 0
	for n < amount {
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
	npcMob, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.respawnTimer = NPC_RESPAWN_DURATION
		return
	}

	// Behavior update
	switch npc.Mode {
		case NPC_MODE_AGGRO: {
			// If the mob is doing something, keep doing it
			if npcMob.Mode != MOB_MODE_IDLE {
				break
			}

			// If the mob is not doing anything, then find a target
			npcRoom := &world.Rooms[npcMob.Data.Room]
			for _, targetHandle := range npcRoom.Occupants {
				// Don't attack yourself
				if targetHandle == npc.mobHandle {
					continue
				}

				// For now, only attack players
				targetMob := world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter == nil {
					continue
				}

				// Found target, set to attack
				npcMob.SetModeAttack(world, npc.mobHandle, targetHandle)
				break
			}
		}
	}
}
