local item = {}

item.name = "Potion of Mana"
item.description = "A blue tonic that gives mana to the drinker."
item.kind = world.ItemKind.CONSUMABLE
item.size = 5

item.on_use = function(user_handle)
    local regen = world.regen_mana(user_handle, 20)
    local user = world.get_mob_data(user_handle, { "name", "room" })
    world.message_room(user.room, string.format("%s drank a %s and regained %d MP.", user.name, item.name, regen))
end

return item
