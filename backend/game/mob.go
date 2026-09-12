package game

import (
	"fmt"
	"log"
	"math/rand/v2"
)

const MOB_MAX_LEVEL int32 = 20
const MOB_EXP_PER_LEVEL int32 = 300

type MobData struct {
	Name string
	Room uint

	Level int32
	Experience int32
	ExperienceToNextLevel int32

	// Base stats
	Vitality int32
	Strength int32
	Agility int32
	Intelligence int32
	Faith int32

	Health int32
	Mana int32

	Spells []Spell
	Inventory ItemList
}

type MobMode int32
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
)

type Mob struct {
	player *Player
	Data MobData

	Mode MobMode
	Target MobHandle
	CastSpell Spell
}

func MobInitFromCharacter(player *Player, character *Character) Mob {
	return Mob {
		player: player,
		Data: character.Data,

		Mode: MOB_MODE_IDLE,
	}
}

func (mob *Mob) IsDead() bool {
	return mob.Data.Health <= 0
}

func (mobData *MobData) GetExpToNextLevel() int32 {
	if mobData.Level == MOB_MAX_LEVEL {
		return 0
	}

	return mobData.Level * MOB_EXP_PER_LEVEL
}

func (mobData *MobData) MaxHealth() int32 {
	return mobData.Vitality * 5
}

func (mobData *MobData) MaxMana() int32 {
	return mobData.Intelligence * 5
}

func (mob *Mob) GrantExperience(experience int32) {
	// This function is only meant for player mobs at this time
	if mob.player == nil {
		return
	}

	for experience > 0 && mob.Data.Level < MOB_MAX_LEVEL {
		if mob.Data.Experience + experience >= mob.Data.ExperienceToNextLevel {
			experience -= mob.Data.ExperienceToNextLevel

			mob.Data.Experience = 0
			mob.Data.ExperienceToNextLevel = mob.Data.GetExpToNextLevel()
			mob.Data.Level++

			*mob.player.inbox <- fmt.Sprintf("Level up! %s is now level %d", mob.Data.Name, mob.Data.Level)

			continue
		}

		mob.Data.Experience += experience
		experience = 0
	}
}


func (mob *Mob) Update(gameState *GameState) {
	switch mob.Mode {
		case MOB_MODE_IDLE:
		case MOB_MODE_ATTACK:
			// Check if target exists
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.IsDead() || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MOB_MODE_IDLE
				break
			}

			room := &gameState.world.Rooms[mob.Data.Room]

			// Check for evasion
			toHitDc := min(0.5, 0.25 * (float32(targetMob.Data.Agility) / float32(mob.Data.Agility)))
			toHitRoll := rand.Float32()
			if toHitRoll < toHitDc {
				room.broadcast(gameState, fmt.Sprintf("%s dodged %s's attack!", targetMob.Data.Name, mob.Data.Name))
				break
			}

			// TODO: Check for crit.

			// Calculate physical damage
			// TODO formula is damage = ((strength / 2) + weapon bonus) - armor
			attackerMinDamage := max(1, mob.Data.Level / 2)
			damage := mob.Data.Strength / 2
			damage = max(damage, attackerMinDamage)
			targetMob.Data.Health -= damage

			// Broadcast result to room
			room.broadcast(gameState, fmt.Sprintf("%s attacked %s for %d damage.", mob.Data.Name, targetMob.Data.Name, damage))
			if targetMob.IsDead() {
				room.broadcast(gameState, fmt.Sprintf("%s has slain %s.", mob.Data.Name, targetMob.Data.Name))
			}
		case MOB_MODE_CAST:
			// Check if target exists
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.Data.Health == 0 || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MOB_MODE_IDLE
				break
			}

			// Check if caster has enough mana
			spellData := SPELL_DATA[mob.CastSpell]
			room := gameState.world.Rooms[mob.Data.Room]
			if mob.Data.Mana < spellData.manaCost {
				room.broadcast(gameState, fmt.Sprintf("%s tried to cast %s, but they don't have enough mana.", mob.Data.Name, spellData.name))
				mob.Mode = MOB_MODE_IDLE
				break
			}

			// Cast spell
			room.broadcast(gameState, fmt.Sprintf("%s cast %s!", mob.Data.Name, spellData.name))
			mob.Data.Mana -= spellData.manaCost
			spellData.onHit(gameState, targetMob)

			mob.Mode = MOB_MODE_IDLE
		default:
			log.Printf("Mob mode %d not handled.", mob.Mode)
	}
}
