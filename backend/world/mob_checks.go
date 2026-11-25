package world

// The kind of pass/fail check a roll is for. Every pass/fail roll goes through rollCheck with its CheckType, so that
// tests can force the result of the next check of a given type with World.PushCheckResult
type CheckType int
const (
	CHECK_TO_HIT CheckType = iota
	CHECK_CRIT
	CHECK_DEFLECT
	CHECK_TAUNT
	CHECK_TYPE_COUNT
)

const CHECK_CHANCE_MIN float32 = 0.05
const CHECK_CHANCE_MAX float32 = 0.95

const CHECK_TO_HIT_WEIGHT float32 = 1.0 / 6.0

func (checkType CheckType) String() string {
	switch checkType {
		case CHECK_TO_HIT:
			return "to-hit"
		case CHECK_CRIT:
			return "crit"
		case CHECK_DEFLECT:
			return "deflect"
		case CHECK_TAUNT:
			return "taunt"
		default:
			return "unknown"
	}
}

// Makes the next check of checkType return result instead of rolling. Forced results for the same type are used in
// the order they were pushed. Only for tests: nothing in the game forces checks
func (world *World) PushCheckResult(checkType CheckType, result bool) {
	world.forcedCheckResults[checkType] = append(world.forcedCheckResults[checkType], result)
}

// How many results pushed with PushCheckResult for checkType haven't been used by a check yet. Only for tests
func (world *World) UnusedCheckResultCount(checkType CheckType) int {
	return len(world.forcedCheckResults[checkType])
}

func contestedCheckChance(rank int32, difficulty int32, weight float32) float32 {
	if rank == 0 && difficulty == 0 {
		return CHECK_CHANCE_MAX
	}
	chance := float32(rank) / (float32(rank) + (float32(difficulty) * weight))
	return max(min(chance, CHECK_CHANCE_MAX), CHECK_CHANCE_MIN)
}

// Passes with the given chance (0 to 1), unless a test forced the result of the next check of checkType
func rollCheck(world *World, checkType CheckType, chance float32) bool {
	// Always roll, so a forced check doesn't change the numbers later rolls get
	result := world.Random.Float32() < chance

	// If there's no forced results, return the result of the roll
	forcedResults := world.forcedCheckResults[checkType]
	if len(forcedResults) == 0 {
		return result
	}

	// Otherwise, return the forced result
	world.forcedCheckResults[checkType] = forcedResults[1:]
	return forcedResults[0]
}

func (mob *Mob) rollTaunt(world *World, targetMob *Mob) bool {
	// Redundant check, should be unnecessary, but better safe than to crash the server
	if !mob.IsPlayer() || !targetMob.IsEvil() {
		return false
	}

	chance := contestedCheckChance(mob.Character.Attribute(ATTR_STR), targetMob.Stat(STAT_FOCUS), 1.0)
	chance *= (1.0 - mob.tauntCooldown)

	return rollCheck(world, CHECK_TAUNT, chance)
}

func (mob *Mob) rollToHit(world *World, defender *Mob) bool {
	chance := contestedCheckChance(mob.Stat(STAT_ACCURACY), defender.Stat(STAT_EVASION), CHECK_TO_HIT_WEIGHT)
	return rollCheck(world, CHECK_TO_HIT, chance)
}

func (mob *Mob) rollCrit(world *World) bool {
	return rollCheck(world, CHECK_CRIT, float32(mob.Stat(STAT_CRIT_CHANCE)) / 100.0)
}

// Rolls whether mob deflects attack, which lowers its damage by mob's deflection amount
func (mob *Mob) rollDeflect(world *World, attack MobAttack) bool {
	deflectChance := (mob.Stat(STAT_DEFLECTION) - attack.penetration) * MOB_DEFLECTION_PER_POINT
	deflectChance = max(min(deflectChance, 100), 0)
	return rollCheck(world, CHECK_DEFLECT, float32(deflectChance) / 100.0)
}
