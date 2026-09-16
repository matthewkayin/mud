package game

// Since each update is 3 seconds, this means it respawns in 30 seconds
// TODO: change this to a longer duration
// TODO: make this customizable per NPC?
const NPC_RESPAWN_DURATION int = 10

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
	npcRoom := &world.Rooms[npcMob.data.Room]
	npcRoom.AddOccupant(npc.mobHandle)

	// Should we broadcast a spawn message to the room here?
}

func (npc *Npc) onMobDeath() {
	npc.respawnTimer = NPC_RESPAWN_DURATION
}

func (npc *Npc) update(world *World) {
	// If respawning, just update timer and don't do anything else
	if npc.respawnTimer != 0 {
		npc.respawnTimer--
		if npc.respawnTimer == 0 {
			npc.spawnMob(world)
		}

		return
	}

	// Check for mob death
	npcMob, npcMobExists := world.Mobs.GetIfExists(npc.mobHandle)
	if !npcMobExists {
		npc.respawnTimer = NPC_RESPAWN_DURATION
		return
	}

	npcRoom := &world.Rooms[npcMob.data.Room]

	// Behavior update
	switch npc.Behavior {
		case NPC_BEHAVIOR_AGGRO:
			// If mob isn't doing anything, find target
			if npcMob.mode == MOB_MODE_IDLE {
				for _, targetHandle := range npcRoom.occupants {
					// Don't attack yourself
					if targetHandle.Equals(npc.mobHandle) {
						continue
					}

					// For now, only attack players
					targetMob := world.Mobs.Get(targetHandle)
					if targetMob.player == nil {
						continue
					}

					// Found target, set to attack
					npcMob.SetModeAttack(targetHandle)
					break
				}
			}
	}
}
