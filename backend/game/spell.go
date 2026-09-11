package game

import (
	"fmt"
)

type Spell int
const (
	SPELL_FIREBOLT = iota
	SPELL_CURE
)

type SpellData struct {
	name string
	description string
	manaCost int

	onHit func(gameState *GameState, target *Mob)
}

var SPELL_DATA = map[Spell]*SpellData {
	SPELL_FIREBOLT: {
		name: "Firebolt",
		description: "Casts a bolt of fire toward the target",
		manaCost: 5,

		onHit: func(gameState *GameState, target *Mob) {
			damage := 10
			target.Data.Health -= damage

			room := gameState.world.Rooms[target.Data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s took %d damage from the firebolt.", target.Data.Name, damage))
			if target.IsDead() {
				room.broadcast(gameState, fmt.Sprintf("%s has burnt to a crisp.", target.Data.Name))
			}
		},
	},
	SPELL_CURE: {
		name: "Cure",
		description: "Heals the target with holy magic",
		manaCost: 5,

		onHit: func(gameState *GameState, target *Mob) {
			healing := min(15, target.Data.MaxHealth() - target.Data.Health)
			target.Data.Health += healing

			room := gameState.world.Rooms[target.Data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s regained %d HP.", target.Data.Name, healing))
		},
	},
}
