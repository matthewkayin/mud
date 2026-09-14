package game

import (
	"fmt"
)

var SPELL_CAST_TIME_INSTANT int32 = 0

type Spell int32
const (
	SPELL_FIREBOLT = iota
	SPELL_CURE
)

type SpellData struct {
	name string
	description string
	manaCost int32
	castTime int32

	onHit func(gameState *GameState, caster *Mob, target *Mob)
}

var SPELL_DATA = map[Spell]*SpellData {
	SPELL_FIREBOLT: {
		name: "Firebolt",
		description: "Casts a bolt of fire toward the target",
		manaCost: 5,
		castTime: 1,

		onHit: func(gameState *GameState, caster *Mob, target *Mob) {
			damage := caster.CalculateMagicDamage(10, target)
			target.Data.Health -= damage

			room := gameState.world.Rooms[target.Data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s took %d damage from the firebolt.", target.Data.Name, damage))
			if target.IsDead() {
				room.broadcast(gameState, fmt.Sprintf("%s has burnt to a crisp.", target.Data.Name))
			} else {
				target.RollForConcentration(gameState, damage)
			}
		},
	},
	SPELL_CURE: {
		name: "Cure",
		description: "Heals the target with holy magic",
		manaCost: 5,
		castTime: SPELL_CAST_TIME_INSTANT,

		onHit: func(gameState *GameState, caster *Mob, target *Mob) {
			healing := caster.CalculateMagicDamage(15, target)
			healingReceived := min(healing, target.Data.MaxHealth() - healing)
			target.Data.Health += healingReceived

			room := gameState.world.Rooms[target.Data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s regained %d HP.", target.Data.Name, healingReceived))
		},
	},
}
