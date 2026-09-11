package game

import (
	"log"
	"fmt"
)

type CharacterData struct {
	Name string
	Room uint

	Health uint
	MaxHealth uint
	Damage uint
}

type MobMode int
const (
	MobModeIdle MobMode = iota
	MobModeAttack
)

type Mob struct {
	player *Player
	Data CharacterData

	Mode MobMode
	Target MobHandle
}

func MobInitFromCharacter(player *Player, character *Character) Mob {
	return Mob {
		player: player,
		Data: character.Data,

		Mode: MobModeIdle,
	}
}

func (mob *Mob) IsDead() bool {
	return mob.Data.Health == 0
}

func (mob *Mob) Update(gameState *GameState) {
	switch mob.Mode {
		case MobModeIdle:
		case MobModeAttack:
			targetMob, targetExists := gameState.world.Mobs.GetIfExists(mob.Target)
			if !targetExists || targetMob.Data.Health == 0 || targetMob.Data.Room != mob.Data.Room {
				mob.Mode = MobModeIdle
				break
			}

			if mob.Data.Damage > targetMob.Data.Health {
				targetMob.Data.Health = 0
			} else {
				targetMob.Data.Health -= mob.Data.Damage
			}

			room := gameState.world.Rooms[mob.Data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s attacked %s for %d damage.", mob.Data.Name, targetMob.Data.Name, mob.Data.Damage))
			if targetMob.Data.Health == 0 {
				room.broadcast(gameState, fmt.Sprintf("%s has slain %s.", mob.Data.Name, targetMob.Data.Name))
			}
		default:
			log.Printf("Mob mode %d not handled.", mob.Mode)
	}
}
