local item = {}

item.name = "Vial of Poison"
item.description = "A vial of dark green poison."
item.kind = world.ItemKind.CONSUMABLE
item.price = 20
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.ITEM

function item.validate_target(user, target)
    local item_kind = world.item_get_kind(target)
    if item_kind ~= world.ItemKind.EQUIPMENT_ONE_HANDED and item_kind ~= world.ItemKind.EQUIPMENT_TWO_HANDED then
        return "That item cannot have poison applied to it."
    end

    return nil
end

function item.on_use(user, target)
    local user_name = world.mob_get_name(user)
    local target_name = world.item_get_name(target)

    world.item_infuse(target, "Poison", world.time_minutes(1), 1.0, 1)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> has applied <r>poison</r> to their <g>%s</g>.", user_name, target_name))
end

item.notes = ""

return item
