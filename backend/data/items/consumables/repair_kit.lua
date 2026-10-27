local item = {}

item.name = "Repair Kit"
item.description = "Materials and tools for maintaing gear in the field."
item.kind = world.ItemKind.CONSUMABLE
item.price = 10
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.ITEM

function item.validate_target(user, target)
    if not world.item_is_repairable(target) then
        return "That item cannot be repaired."
    end
    if world.item_get_durability_percent(target) >= 100 then
        return "That item is not damaged."
    end

    return nil
end

function item.on_use(user, target)
    local user_name = world.mob_get_name(user)
    local target_name = world.item_get_name(target)

    world.item_repair(target)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> repaired their <g>%s</g>.", user_name, target_name))
end

item.notes = ""

return item
