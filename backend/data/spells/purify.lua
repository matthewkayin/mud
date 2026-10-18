local spell = {}

spell.name = "Purify"
spell.description = "Cures the target of any ailments"
spell.casts_to_learn = 1

spell.mana_cost = 5
spell.cast_time = 1
spell.targets = world.SpellTarget.ALLY

function spell.on_cast(caster, target)
    local room = world.mob_get_room(caster)
    local target_name = world.mob_get_name(target)

    local index = 1
    while index <= world.mob_get_condition_count(target) do
        if world.mob_condition_is_ailment(target, index) then
            local condition_name = world.mob_get_condition_name(target, index)
            world.mob_remove_condition(target, condition_name)
            world.message_room(room, string.format("<y>%s</y> was cured of <r>%s</r>!", target_name, condition_name))
        else
            index = index + 1
        end
    end
end

return spell
