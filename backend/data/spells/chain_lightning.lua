local spell = {}

spell.name = "Chain Lightning"
spell.description = "Casts a bolt of lightning that strikes every enemy in the room"
spell.casts_to_learn = 75

spell.mana_cost = 20
spell.cast_time = 2
spell.target_type = world.TargetType.NONE

function spell.on_cast(caster)
    local room = world.mob_get_room(caster)
    local caster_name = world.mob_get_name(caster)
    world.message_room(room, string.format("Bolts of lightning arced from <y>%s's</y> fingertips!", caster_name))

    for occupant_index = 1,world.room_get_occupant_count(room) do
        local occupant = world.room_get_occupant(room, occupant_index)
        if world.mobs_are_enemies(caster, occupant) then
            local damage = world.deal_magic_damage(caster, occupant, 30)
            local occupant_name = world.mob_get_name(occupant)
            world.message_room(room, string.format("<y>%s</y> took <r>%d damage</r> from the lightning.", occupant_name, damage))
            if world.mob_get_health(occupant) <= 0 then
                world.message_room(room, string.format("<y>%s</y> has been <r>reduced to ash</r>.", occupant_name))
            end
        end
    end
end

return spell
