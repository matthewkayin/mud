local item = {}

item.name = "Wormy Apple"
item.description = "This apple is dimpled with little holes where worms have had their breakfasts."
item.kind = world.ItemKind.CONSUMABLE
item.price = 1
item.size = 1

item.usable_by_job = "any"
item.target_type = world.TargetType.NONE

function item.on_use(user)
    local healing = world.heal(user, 5)
    local user_name = world.mob_get_name(user)
    world.message_room(world.mob_get_room(user), string.format("<y>%s</y> ate an apple <g>%s</g> and regained <g>%d HP</g>.", user_name, item.name, healing))

    local makes_player_sick = math.random(5) == 1
    if makes_player_sick then
        local condition_name = "Poison"
        world.mob_add_condition(user, condition_name, world.time_minutes(1))
        world.message_player(user, string.format("That apple didn't sit so well with you. You have contracted <r>%s</r>.", condition_name))
    end
end

return item
