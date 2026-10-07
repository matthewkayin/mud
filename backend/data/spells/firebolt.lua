local spell = {}

spell.Name = "Firebolt"
spell.Description = "Casts a bolt of fire toward the target"
spell.CastsToLearn = 50

spell.ManaCost = 5
spell.CastTime = 1
spell.CanTargetPlayers = false

spell.OnHit = function(caster, target)
	world.log("{target} took {damage} damage from the firebolt", {
		target = "Friend",
		damage = 5,
	})
end

return spell
