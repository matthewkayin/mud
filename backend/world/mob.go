package world

import (
	"fmt"
)

const MOB_PLAYER_NONE = -1

type MobMode int32
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
	MOB_MODE_USE_ITEM
)

type Mob struct {
	PlayerCharacter *Character
	Data MobData

	mode MobMode
	target MobHandle

	castSpell Spell
	castTimer int32
	useItemId ItemId
}

func MobInit(data *MobData) Mob {
	mob := Mob {
		PlayerCharacter: nil,
		Data: *data,
		mode: MOB_MODE_IDLE,
	}

	mob.Data.Equipment.CalculateStatBonuses()
	return mob
}

func MobInitFromCharacter(character *Character) Mob {
	mob := MobInit(&character.Data)
	mob.PlayerCharacter = character

	return mob
}

func (mob *Mob) GrantExperience(world *World, experience int32) {
	// This function is only meant for player mobs at this time
	if mob.PlayerCharacter == nil {
		return
	}

	for experience > 0 && mob.Data.Level < MOB_MAX_LEVEL {
		if mob.Data.Experience + experience >= mob.Data.ExperienceToNextLevel {
			experience -= mob.Data.ExperienceToNextLevel

			// Increase level
			mob.Data.Experience = 0
			mob.Data.ExperienceToNextLevel = mob.Data.GetExpToNextLevel()
			mob.Data.Level++

			// Recalculate stats
			mob.PlayerCharacter.recalculateStats()
			mob.Data.Stats = mob.PlayerCharacter.Data.Stats

			// Announce level up message
			world.messagePlayer(mob.PlayerCharacter.PlayerId, fmt.Sprintf("Level up! %s is now level %d.", mob.Data.Name, mob.Data.Level))
			continue
		}

		mob.Data.Experience += experience
		experience = 0
	}
}

func (mob *Mob) SetModeAttack(world *World, mobHandle MobHandle, targetHandle MobHandle) {
	mob.mode = MOB_MODE_ATTACK
	mob.target = targetHandle

	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_SET_TARGET,
		Data: EventMobSetTarget {
			Attacker: mobHandle,
			Defender: targetHandle,
		},
	})
}

func (mob *Mob) SetModeCast(world *World, mobHandle MobHandle, spell Spell, targetHandle MobHandle) {
	mob.mode = MOB_MODE_CAST
	mob.target = targetHandle
	mob.castSpell = spell
	mob.castTimer = SPELL_DATA[spell].castTime

	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_SET_TARGET,
		Data: EventMobSetTarget {
			Attacker: mobHandle,
			Defender: targetHandle,
		},
	})
}

func (mob *Mob) SetModeUseItem(world *World, mobHandle MobHandle, itemId ItemId, targetHandle MobHandle) {
	mob.mode = MOB_MODE_USE_ITEM
	mob.target = targetHandle
	mob.useItemId = itemId

	world.pushEvent(Event {
		EventType: EVENT_TYPE_MOB_SET_TARGET,
		Data: EventMobSetTarget {
			Attacker: mobHandle,
			Defender: targetHandle,
		},
	})
}
