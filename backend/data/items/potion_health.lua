local item = {}

item.name = "Potion of Health"
item.description = "A red tonic that gives health to the drinker"
item.kind = world.ItemKind.CONSUMABLE
item.size = 5

item.on_use = function(user_handle)
    local healing = world.heal(user_handle, 20)
    local user = world.get_mob_data(user_handle, { "name", "room" })
    world.message_room(user.room, string.format("%s drank a %s and regained %d HP.", user.name, item.name, healing))
end

return item
