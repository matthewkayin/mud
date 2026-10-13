local spell = {}

spell.name = "Blessing of Might"
spell.description = "Grants increased power to you and your allies for a brief time"
spell.casts_to_learn = 1

spell.mana_cost = 5
spell.cast_time = 1
spell.can_target_players = true

-- TODO: make this an AOE
function spell.on_hit(caster, target)
    local room = world.mob_get_room(caster)
    local caster_name = world.mob_get_name(caster)
    local target_name = world.mob_get_name(target)

    local duration = world.mob_calc_magic_condition_duration(caster, world.time_minutes(1))
    world.mob_add_condition(target, world.ConditionType.BLESSING_OF_MIGHT, duration)
    world.message_room(room, string.format("%s gave %s a Blessing of Might!", caster_name, target_name))
end

return spell
