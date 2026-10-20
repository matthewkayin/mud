local spell = {}

spell.name = "Firebolt"
spell.description = "Casts a bolt of fire toward the target"
spell.casts_to_learn = 50

spell.mana_cost = 5
spell.cast_time = 1
spell.target_type = world.TargetType.ENEMY

function spell.on_cast(caster, target)
    local damage = world.deal_magic_damage(caster, target, 10)

    local room = world.mob_get_room(caster)
    local target_name = world.mob_get_name(target)
    world.message_room(room, string.format("<y>%s</y> took <r>%d damage</r> from the firebolt.", target_name, damage))
    if world.mob_get_health(target) <= 0 then
        world.message_room(room, string.format("<y>%s</y> has <r>burnt to a crisp</r>.", target_name))
    end
end

return spell
