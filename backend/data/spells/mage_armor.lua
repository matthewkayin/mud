local spell = {}

spell.name = "Mage Armor"
spell.description = "Defend yourself with a magical enchantment"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = world.SPELL_CAST_TIME_INSTANT
spell.targets = world.SpellTarget.NONE

function spell.on_cast(caster)
    local caster_name = world.mob_get_name(caster)
    local duration = world.mob_calc_magic_condition_duration(caster, world.time_minutes(1))
    local room = world.mob_get_room(caster)

    world.mob_add_condition(caster, "Mage Armor", duration)

    world.message_room(room, string.format("<y>%s</y> has shrouded themself in powerful defensive magicks!", caster_name))
end

return spell
