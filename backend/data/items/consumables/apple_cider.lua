local item = {}

item.name = "Apple Cider"
item.description = "An amber beverage made from pressed apples."
item.kind = world.ItemKind.CONSUMABLE
item.price = 5
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.NONE

item.notes = ""

function item.on_use(user)
    local healing = world.heal(user, 10)
    local user_name = world.mob_get_name(user)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> drank an <g>%s</g> and regained <g>%d HP</g>. Refreshing!", user_name, item.name, healing))
end

return item
