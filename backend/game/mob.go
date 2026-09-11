package game

import (
	"log"
	"fmt"
)

type MobData struct {
	Name string
	Room uint

	// Base stats
	Vigor int
	Strength int
	Agility int
	Intelligence int
	Faith int

	Health int
}

type MobMode int
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
)

type Mob struct {
	player *Player
	Data MobData

	Mode MobMode
	Target MobHandle
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
	return mobData.Vigor * 5
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
			if targetMob.Data.Health == 0 {
				room.broadcast(gameState, fmt.Sprintf("%s has slain %s.", mob.Data.Name, targetMob.Data.Name))
			}
		default:
			log.Printf("Mob mode %d not handled.", mob.Mode)
	}
}
