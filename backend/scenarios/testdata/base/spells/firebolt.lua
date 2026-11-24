local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire at the target."
spell.required_skill = world.Skill.EVOCATION

spell.mana_cost = 4
spell.cast_time = world.SPELL_CAST_TIME_INSTANT
spell.target_type = world.TargetType.ENEMY

function spell.on_cast(caster, target)
	local damage = world.mob_magic_damage(caster, target, 20)

	local room = world.mob_get_room(caster)
	local caster_name = world.mob_get_name(caster)
	local target_name = world.mob_get_name(target)

	world.message_room(room, string.format("A bolt of fire flew from <y>%s's</y> hands!", caster_name))
	world.message_room(room, string.format("<y>%s</y> took <r>%d damage</r>.", target_name, damage))
	if world.mob_is_dead(target) then
		world.message_room(room, string.format("<y>%s</y> has <r>burnt to a crisp</r>.", target_name))
	end
end

return spell
