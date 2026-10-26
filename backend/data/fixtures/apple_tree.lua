local fixture = {}

fixture.name = "Apple Tree"
fixture.description = "A tall tree with lots of apples on it. The apples are out of reach."

fixture.interactions = {
    kick = function(instance, player)
        local apple_roll = math.random(3)
        if apple_roll == 1 then
            local room = world.mob_get_room(player)
            world.message_player(player, "You kicked the Apple Tree... and an <g>Apple</g> fell from the tree!")
            world.room_add_item(room, "Apple", 1)
        elseif apple_roll == 2 then
            local room = world.mob_get_room(player)
            world.message_player(player, "You kicked the Apple Tree... and a <g>Wormy Apple</g> fell from the tree!")
            world.room_add_item(room, "Wormy Apple", 1)
        else
            world.message_player(player, "You kicked the Apple Tree... but nothing happened.")
        end
    end
}

return fixture
