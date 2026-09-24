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
	Name string
	Description string
	CastsToLearn int32

	ManaCost int32
	CastTime int32
	CanTargetPlayers bool

	onHit func(world *World, caster *Mob, target *Mob)
}

var SPELL_DATA = []*SpellData {
	SPELL_FIREBOLT: {
		Name: "Firebolt",
		Description: "Casts a bolt of fire toward the target",
		CastsToLearn: 50,

		ManaCost: 5,
		CastTime: 1,
		CanTargetPlayers: false,

		onHit: func(world *World, caster *Mob, target *Mob) {
			damage := caster.calculateMagicDamage(10, target)
			target.Data.Health -= damage

			world.messageRoom(target.Data.Room, fmt.Sprintf("%s took %d damage from the firebolt.", target.GetName(), damage))
			if target.IsDead() {
				world.messageRoom(target.Data.Room, fmt.Sprintf("%s has burnt to a crisp.", target.GetName()))
			} else {
				target.rollForConcentration(world, damage)
			}
		},
	},
	SPELL_CURE: {
		Name: "Cure",
		Description: "Heals the target with holy magic",
		CastsToLearn: 50,

		ManaCost: 5,
		CastTime: SPELL_CAST_TIME_INSTANT,
		CanTargetPlayers: true,

		onHit: func(world *World, caster *Mob, target *Mob) {
			healing := caster.calculateMagicDamage(15, target)
			healingReceived := min(healing, target.Data.MaxHealth() - target.Data.Health)
			target.Data.Health += healingReceived

			world.messageRoom(target.Data.Room, fmt.Sprintf("%s regained %d HP.", target.GetName(), healingReceived))
		},
	},
}
