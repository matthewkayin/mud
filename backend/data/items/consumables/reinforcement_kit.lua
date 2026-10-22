local item = {}

item.name = "Reinforcement Kit"
item.description = "Materials and tools for upgrading weapons. While durability on a weapon is above its normal maximum it does additional damage."
item.kind = world.ItemKind.CONSUMABLE
item.price = 100
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.INVENTORY_ITEM

item.on_use = function(user, target)
    local user_name = world.mob_get_name(user)
    local target_name = world.mob_get_inventory_item_name(user, target)

    local success = false

    success = world.mob_reinforce_item(user, target)


    if success == false then
        world.message_player(user, "That is not a valid item.")
    else
        world.message_room(world.mob_get_room(user), string.format("<y>%s</y> used a <g>%s</g> on their <g>%s</g>.", user_name, item.name, target_name))
    end
end

item.notes = ""

return item
