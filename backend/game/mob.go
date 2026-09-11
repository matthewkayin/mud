package game

import (
	"log"
	"fmt"
)

type MobData struct {
	Name string
	Room uint

	// Base stats
	Vitality int
	Strength int
	Agility int
	Intelligence int
	Faith int

	Health int
	Mana int

	Spells []Spell
	Inventory ItemList
}

type MobMode int
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

func (mobData *MobData) MaxHealth() int {
	return mobData.Vitality * 5
}

func (mobData *MobData) MaxMana() int {
	return mobData.Intelligence * 5
}

func (mob *Mob) AttackDamage() int {
	return mob.Data.Strength
}

func (mob *Mob) Update(gameState *GameState) {
	switch mob.Mode {
		case MOB_MODE_IDLE:
		case MOB_MODE_ATTACK:
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.Data.Health == 0 || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MOB_MODE_IDLE
				break
			}

			damage := mob.AttackDamage()
			targetMob.Data.Health -= damage

			room := gameState.world.Rooms[mob.Data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s attacked %s for %d damage.", mob.Data.Name, targetMob.Data.Name, damage))
			if targetMob.IsDead() {
				room.broadcast(gameState, fmt.Sprintf("%s has slain %s.", mob.Data.Name, targetMob.Data.Name))
			}
		case MOB_MODE_CAST:
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.Data.Health == 0 || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MOB_MODE_IDLE
				break
			}

			spellData := SPELL_DATA[mob.CastSpell]
			room := gameState.world.Rooms[mob.Data.Room]
			if mob.Data.Mana < spellData.manaCost {
				room.broadcast(gameState, fmt.Sprintf("%s tried to cast %s, but they don't have enough mana.", mob.Data.Name, spellData.name))
				mob.Mode = MOB_MODE_IDLE
				break
			}

			room.broadcast(gameState, fmt.Sprintf("%s cast %s!", mob.Data.Name, spellData.name))
			mob.Data.Mana -= spellData.manaCost
			spellData.onHit(gameState, targetMob)

			mob.Mode = MOB_MODE_IDLE
		default:
			log.Printf("Mob mode %d not handled.", mob.Mode)
	}
}
