local item = {}

item.name = "Weapon Reinforcement Kit"
item.description = "Materials and tools for upgrading weapons. While durability on a weapon is above its normal maximum it does additional damage."
item.kind = world.ItemKind.CONSUMABLE
item.price = 100
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.INVENTORY_ITEM

item.on_use = function(user, target)
    local reinforcement = 50
    local user_name = world.mob_get_name(user)
    local target_name = world.mob_get_inventory_item_name(user, target)

    local success = false
    local kind = world.mob_get_inventory_item_kind(user, target)

    if kind == "One-Handed Weapon" or kind == "Two-Handed Weapon" then
        world.message_room(world.mob_get_room(user), "!!!this is a weapon")
        success = world.mob_reinforce_item(user, target, reinforcement)
    end

    if success == false then
        world.message_player(user, "That is not a valid item.")
    else
        world.message_room(world.mob_get_room(user), string.format("<y>%s</y> used a <g>%s</g> on their <g>%s</g>.", user_name, item.name, target_name))
    end
end

item.notes = ""

return item
