local spell = {}

spell.name = "Unlock"
spell.description = "Unlocks a physical or magical lock"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = world.SPELL_CAST_TIME_INSTANT
spell.targets = world.SpellTarget.ARGUMENT

function spell.validate_argument(caster, argument)
    local room = world.mob_get_room(caster)
    local directions = {
        world.Direction.NORTH,
        world.Direction.EAST,
        world.Direction.SOUTH,
        world.Direction.WEST
    }

    for index = 1, #directions do
        if argument == directions[index] then
            local exit_room = world.room_get_exit(room, directions[index])
            if exit_room == world.ROOM_NONE then
                return string.format("This room has no %s exit.", directions[index])
            end

            if not world.room_is_exit_locked(room, directions[index]) then
                return string.format("The %s exit is not locked.", directions[index])
            end

            return nil
        end
    end

    return string.format("'%s' is not a valid direction. The directions are 'north', 'east', 'south', and 'west'.", argument)
end

function spell.on_cast(caster, argument)
    local room = world.mob_get_room(caster)
    local caster_name = world.mob_get_name(caster)
    world.room_set_exit_locked(room, argument, false)
    world.message_room(room, string.format("<y>%s</y> unlocked the %s exit!", caster_name, argument))
end

return spell
