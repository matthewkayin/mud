local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire toward the target"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = 1
spell.can_target_players = false

function spell.on_hit(caster_handle, target_handle)
    local damage = world.deal_magic_damage(caster_handle, target_handle, 10)

    local room = world.mob_get_room(caster_handle)
    local target_name = world.mob_get_name(target_handle)
    world.message_room(room, string.format("%s took %d damage from the firebolt.", target_name, damage))
    if world.mob_get_health(target_handle) <= 0 then
        world.message_room(room, string.format("%s has burnt to a crisp.", target_name))
    end
end

return spell
