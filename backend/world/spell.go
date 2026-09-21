package world

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
			damage := caster.CalculateMagicDamage(10, target)
			target.Data.Health -= damage

			room := world.Rooms[target.data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s took %d damage from the firebolt.", target.data.Name, damage))
			if target.IsDead() {
				room.broadcast(gameState, fmt.Sprintf("%s has burnt to a crisp.", target.data.Name))
			} else {
				target.RollForConcentration(gameState, damage)
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

		onHit: func(gameState *GameState, caster *Mob, target *Mob) {
			healing := caster.CalculateMagicDamage(15, target)
			healingReceived := min(healing, target.data.MaxHealth() - target.data.Health)
			target.data.Health += healingReceived

			room := gameState.world.Rooms[target.data.Room]
			room.broadcast(gameState, fmt.Sprintf("%s regained %d HP.", target.data.Name, healingReceived))
		},
	},
}
