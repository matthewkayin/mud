package world

import (
	"math/rand/v2"
)

const CHECK_CHANCE_MIN float32 = 0.05
const CHECK_CHANCE_MAX float32 = 0.95

func contestedCheckChance(rank int32, difficulty int32, weight float32) float32 {
	if rank == 0 && difficulty == 0 {
		return CHECK_CHANCE_MAX
	}
	chance := float32(rank) / (float32(rank) + (float32(difficulty) * weight))
	return max(min(chance, CHECK_CHANCE_MAX), CHECK_CHANCE_MIN)
}

func rollCheck(chance float32) bool {
	return rand.Float32() < chance
}

func (mob *Mob) rollTaunt(targetMob *Mob) bool {
	// Redundant check, should be unnecessary, but better safe than to crash the server
	if !mob.IsPlayer() || !targetMob.IsEvil() {
		return false
	}

	chance := contestedCheckChance(mob.Character.Attribute(ATTR_STR), targetMob.Stat(STAT_FOCUS), 1.0)
	chance *= (1.0 - mob.tauntCooldown)

	return rollCheck(chance)
}
