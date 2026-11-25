local spell = {}

spell.name = "Heal"
spell.description = "Heals the target."
spell.required_skill = world.Skill.DRUIDRY

spell.mana_cost = 4
spell.cast_time = world.SPELL_CAST_TIME_INSTANT
spell.target_type = world.TargetType.ALLY

function spell.on_cast(caster, target)
    local healing = world.mob_magic_heal(caster, target, 10)

	local room = world.mob_get_room(caster)
	local target_name = world.mob_get_name(target)

	world.message_room(room, string.format("<y>%s</y> regained <r>%d HP</r>.", target_name, healing))
end

return spell
