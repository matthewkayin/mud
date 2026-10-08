local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire toward the target"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = 1
spell.can_target_players = false

spell.on_hit = function(caster, target)
	world.log("{target} took {damage} damage from the firebolt", {
		target = "Friend",
		damage = 5,
	})
end

return spell
