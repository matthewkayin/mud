package world

const CHECK_CHANCE_MIN float32 = 0.05
const CHECK_CHANCE_MAX float32 = 0.95

const CHECK_TO_HIT_WEIGHT float32 = 1.0 / 6.0

func contestedCheckChance(rank int32, difficulty int32, weight float32) float32 {
	if rank == 0 && difficulty == 0 {
		return CHECK_CHANCE_MAX
	}
	chance := float32(rank) / (float32(rank) + (float32(difficulty) * weight))
	return max(min(chance, CHECK_CHANCE_MAX), CHECK_CHANCE_MIN)
}

func rollCheck(world *World, chance float32) bool {
	return world.Random.Float32() < chance
}

func (mob *Mob) rollTaunt(world *World, targetMob *Mob) bool {
	// Redundant check, should be unnecessary, but better safe than to crash the server
	if !mob.IsPlayer() || !targetMob.IsEvil() {
		return false
	}

	chance := contestedCheckChance(mob.Character.Attribute(ATTR_STR), targetMob.Stat(STAT_FOCUS), 1.0)
	chance *= (1.0 - mob.tauntCooldown)

	return rollCheck(world, chance)
}

func (mob *Mob) rollToHit(world *World, defender *Mob) bool {
	chance := contestedCheckChance(mob.Stat(STAT_ACCURACY), defender.Stat(STAT_EVASION), CHECK_TO_HIT_WEIGHT)
	return rollCheck(world, chance)
}

func (mob *Mob) rollCrit(world *World) bool {
	return rollCheck(world, float32(mob.Stat(STAT_CRIT_CHANCE)) / 100.0)
}
