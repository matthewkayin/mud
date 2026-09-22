package world

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
	castsToLearn int32

	manaCost int32
	castTime int32
	canTargetPlayers bool

	onHit func(world *World, caster *Mob, target *Mob)
}

var SPELL_DATA = []*SpellData {
	SPELL_FIREBOLT: {
		name: "Firebolt",
		description: "Casts a bolt of fire toward the target",
		castsToLearn: 50,

		manaCost: 5,
		castTime: 1,
		canTargetPlayers: false,

		onHit: func(world *World, caster *Mob, target *Mob) {
			damage := caster.calculateMagicDamage(10, target)
			target.Data.Health -= damage

			world.messageRoom(target.Data.Room, fmt.Sprintf("%s took %d damage from the firebolt.", target.Data.Name, damage))
			if target.IsDead() {
				world.messageRoom(target.Data.Room, fmt.Sprintf("%s has burnt to a crisp.", target.Data.Name))
			} else {
				target.rollForConcentration(world, damage)
			}
		},
	},
	SPELL_CURE: {
		name: "Cure",
		description: "Heals the target with holy magic",
		castsToLearn: 50,

		manaCost: 5,
		castTime: SPELL_CAST_TIME_INSTANT,
		canTargetPlayers: true,

		onHit: func(world *World, caster *Mob, target *Mob) {
			healing := caster.calculateMagicDamage(15, target)
			healingReceived := min(healing, target.Data.MaxHealth() - target.Data.Health)
			target.Data.Health += healingReceived

			world.messageRoom(target.Data.Room, fmt.Sprintf("%s regained %d HP.", target.Data.Name, healingReceived))
		},
	},
}
