local spell = {}

spell.name = "Cure"
spell.description = "Heals the target with holy magic"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = world.SPELL_CAST_TIME_INSTANT
spell.can_target_players = true

function spell.on_hit(caster_handle, target_handle)
    local healing = world.magic_heal(caster_handle, target_handle, 15)

    local target = world.get_mob_data(target_handle, { "room", "name", })
    world.message_room(target.room, string.format("%s regained %d HP.", target.name, healing))
end

return spell
