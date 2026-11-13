package world

const NPC_RESPAWN_TIMER_NOT_STARTED int = -1
const NPC_RESPAWN_DURATION int = (5 * 60) / WORLD_SECONDS_PER_UPDATE

type NpcSpawner struct {
	npcData *NpcData
	mobHandle MobHandle
	respawnDuration int
	respawnTimer int
}

func (spawner *NpcSpawner) update(world *World, roomIndex int) {
	// If the timer has not started, it means we should have a mob
	if spawner.respawnTimer == NPC_RESPAWN_TIMER_NOT_STARTED {
		// Check to see if our mob is still alive
		_, exists := world.Mobs.GetIfExists(spawner.mobHandle)
		if !exists {
			spawner.respawnTimer = spawner.respawnDuration
		}
	}

	// If the timer has started, count it down
	if spawner.respawnTimer != NPC_RESPAWN_TIMER_NOT_STARTED {
		spawner.respawnTimer--
		if spawner.respawnTimer <= 0 {
			spawner.mobHandle = world.spawnNpc(NpcSpawnParams {
				data: spawner.npcData,
				room: roomIndex,
				movementType: NPC_MOVEMENT_TYPE_STATIONARY,
			})
			spawner.respawnTimer = NPC_RESPAWN_TIMER_NOT_STARTED
		}
	}
}
