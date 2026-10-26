local item = {}

item.name = "Antidote"
item.description = "A clear elixir that cures basic ailments like conventional poisons and diseases."
item.kind = world.ItemKind.CONSUMABLE
item.price = 10
item.size = 5

item.usable_by_job = "any"
item.target_type = world.TargetType.NONE

function item.on_use(user)
    world.mob_remove_ailments(user)
    local user_name = world.mob_get_name(user)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> drank an <g>%s</g>.", user_name, item.name))
end

item.notes = ""

return item
