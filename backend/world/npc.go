package world

import (
	"fmt"
)

// TODO: change this to a longer duration
// TODO: make this customizable per NPC?
const NPC_RESPAWN_DURATION int = 60 / WORLD_SECONDS_PER_UPDATE

type NpcBehavior int
const (
	NPC_BEHAVIOR_AGGRO = iota
)

type Npc struct {
	Behavior NpcBehavior
	Data MobData

	mobHandle MobHandle
	respawnTimer int
}

func (npc *Npc) init(world *World) {
	npc.spawnMob(world)
}

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
	switch npc.Behavior {
		case NPC_BEHAVIOR_AGGRO: {
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
