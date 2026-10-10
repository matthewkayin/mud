local spell = {}

spell.name = "Cure"
spell.description = "Heals the target with holy magic"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = world.SPELL_CAST_TIME_INSTANT
spell.can_target_players = true

function spell.on_hit(caster_handle, target_handle)
    local healing = world.magic_heal(caster_handle, target_handle, 15)

    local target_name = world.mob_get_name(target_handle)
    world.message_room(world.mob_get_room(target_handle), string.format("%s regained %d HP.", target_name, healing))
end

return spell
