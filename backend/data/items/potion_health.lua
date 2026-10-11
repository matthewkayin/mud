local item = {}

item.name = "Potion of Health"
item.description = "A red tonic that gives health to the drinker"
item.kind = world.ItemKind.CONSUMABLE
item.size = 5

item.on_use = function(user)
    local healing = world.heal(user, 20)
    local user_name = world.mob_get_name(user)
    world.message_room(world.mob_get_room(user), string.format("%s drank a %s and regained %d HP.", user_name, item.name, healing))
end

return item
