local item = {}

item.name = "Potion of Mana"
item.description = "A blue tonic that gives mana to the drinker."
item.kind = world.ItemKind.CONSUMABLE
item.size = 5

item.on_use = function(user_handle)
    local regen = world.regen_mana(user_handle, 20)
    local user_name = world.mob_get_name(user_handle)
    world.message_room(world.mob_get_room(user_handle), string.format("%s drank a %s and regained %d MP.", user_name, item.name, regen))
end

return item
