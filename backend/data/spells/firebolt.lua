local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire toward the target"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = 1
spell.can_target_players = false

function spell.on_hit(caster_handle, target_handle)
    local damage = world.deal_magic_damage(caster_handle, target_handle, 10)

    local caster = world.get_mob_data(caster_handle, { "room" })
    local target = world.get_mob_data(target_handle, { "name", "health" })
    world.message_room(caster.room, "{target} took {damage} damage from the firebolt.", {
        target = target.name,
        damage = damage,
    })
    if target.health <= 0 then
        world.message_room(caster.room, "{target} has burnt to a crisp.", {
            target = target.name
        })
    end
end

return spell
