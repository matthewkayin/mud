local spell = {}

spell.name = "Blessing of Might"
spell.description = "Grants increased power to you and your allies for a brief time"
spell.casts_to_learn = 1

spell.mana_cost = 5
spell.cast_time = 1
spell.target_type = world.TargetType.NONE

function spell.on_cast(caster)
    local room = world.mob_get_room(caster)
    local caster_name = world.mob_get_name(caster)

    local duration = world.mob_calc_magic_condition_duration(caster, world.time_minutes(1))

    for index = 1, world.room_get_occupant_count(room) do
        local occupant = world.room_get_occupant(room, index)
        if world.mobs_are_allies(caster, occupant) then
            world.mob_add_condition(occupant, "Blessing of Might", duration)
        end
    end

    world.message_room(room, string.format("<y>%s</y> gave the party a <g>Blessing of Might</g>!", caster_name))
end

return spell
