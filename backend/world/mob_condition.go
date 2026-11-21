package world

import (
	"log"
)

type MobCondition int
const (
	MOB_CONDITION_UNHARMED MobCondition = iota
	MOB_CONDITION_SCRATCHED
	MOB_CONDITION_HURT
	MOB_CONDITION_WOUNDED
	MOB_CONDITION_BADLY_WOUNDED
	MOB_CONDITION_NEAR_DEATH
)

func (condition MobCondition) String() string {
	switch condition {
		case MOB_CONDITION_UNHARMED:
			return "Unharmed"
		case MOB_CONDITION_SCRATCHED:
			return "Scratched"
		case MOB_CONDITION_HURT:
			return "Hurt"
		case MOB_CONDITION_WOUNDED:
			return "Wounded"
		case MOB_CONDITION_BADLY_WOUNDED:
			return "Badly Wounded"
		case MOB_CONDITION_NEAR_DEATH:
			return "Near Death"
		default:
			log.Printf("warn: unhandled mob condition %d", condition)
			return ""
	}
}

func (mob *Mob) condition() MobCondition {
	healthPercent := (mob.Health * 100) / mob.Stat(STAT_MAX_HEALTH)
	if healthPercent == 100 {
		return MOB_CONDITION_UNHARMED
	} else if healthPercent >= 80 {
		return MOB_CONDITION_SCRATCHED
	} else if healthPercent >= 50 {
		return MOB_CONDITION_HURT
	} else if healthPercent >= 25 {
		return MOB_CONDITION_WOUNDED
	} else if healthPercent >= 10 {
		return MOB_CONDITION_BADLY_WOUNDED
	} else {
		return MOB_CONDITION_NEAR_DEATH
	}
}
