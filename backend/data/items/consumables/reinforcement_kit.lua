local item = {}

item.name = "Reinforcement Kit"
item.description = "Materials and tools for upgrading weapons. While durability on a weapon is above its normal maximum it does additional damage."
item.kind = world.ItemKind.CONSUMABLE
item.price = 100
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.INVENTORY_ITEM

function item.validate_target(user, target)
    if not world.mob_item_is_repairable(user, target) then
        return "That item cannot be repaired."
    end
    if world.mob_get_inventory_item_durability_percent(user, target) >= world.ITEM_REINFORCED_DURABILITY_PERCENT then
        return "That item is already reinforced."
    end

    return nil
end

function item.on_use(user, target)
    local user_name = world.mob_get_name(user)
    local target_name = world.mob_get_inventory_item_name(user, target)

    world.mob_reinforce_item(user, target)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> reinforced their <g>%s</g>.", user_name, target_name))
end

item.notes = ""

return item
