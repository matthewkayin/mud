local item = {}

item.name = "Potion of Mana"
item.description = "A blue tonic that gives mana to the drinker."
item.kind = world.ItemKind.CONSUMABLE
item.price = 20
item.size = 5

item.usable_by_job = "any"
item.target_type = world.TargetType.NONE

item.on_use = function(user)
    local regen = world.regen_mana(user, 20)
    local user_name = world.mob_get_name(user)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> drank a <g>%s</g> and regained %d MP.", user_name, item.name, regen))
end

item.notes = ""

return item
