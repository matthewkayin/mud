local item = {}

item.name = "Apple"
item.description = "A red piece of fruit, known to deter physicians."
item.kind = world.ItemKind.CONSUMABLE
item.price = 2
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.NONE

item.notes = ""

function item.on_use(user)
    local healing = world.heal(user, 5)
    local user_name = world.mob_get_name(user)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> ate an apple <g>%s</g> and regained <g>%d HP</g>.", user_name, item.name, healing))
end

return item
