local spell = {}

spell.name = "Chain Lightning"
spell.description = "Casts a chain of lightning against all targets in the room"
spell.casts_to_learn = 50

spell.mana_cost = 12
spell.cast_time = 2
spell.targets = world.SpellTarget.NONE

function spell.on_cast(caster)
    local room = world.mob_get_room(caster)
    local caster_name = world.mob_get_name(caster)

    world.message_room(room, string.format("Lightning sparks from <y>%s's</y> fingertips!", caster_name))

    -- Collect targets first so the loop doesn't depend on the room's occupant list
    local targets = {}
    for index = 1, world.room_get_occupant_count(room) do
        local occupant = world.room_get_occupant(room, index)
        if world.mob_is_npc(occupant) and not world.mob_is_dead(occupant) then
            table.insert(targets, occupant)
        end
    end

    for index = 1, #targets do
        local target = targets[index]
        local damage = world.deal_magic_damage(caster, target, 7)
        local target_name = world.mob_get_name(target)
        world.message_room(room, string.format("<y>%s</y> took <r>%d damage</r> from the lightning!", target_name, damage))
        if world.mob_is_dead(target) then
            world.message_room(room, string.format("<y>%s</y> has been <r>reduced to a pile of ash</r>!", target_name))
        end
    end
end

return spell
