local room = {}

room.name = "Kitchen"
room.description = "In this room, little explosions are drawn with colored masking tape. There is a long, rectangular dining table, lined with chairs. Along the countertop is an assortment of too-many coffee makers."
room.details = {
    ["coffee makers"] = "Aside from the standard black-and-decker sludge maker, there are an assortment of coffee makers left behind by the opinionated drinkers of RC's past. A chemex, aeropress, hell, even an espresso machine are all available for you to use.",
    ["espresso machine"] = "It's a Breville Barista Express. You didn't expect a commercial one, did you?"
}
room.position = { x = -1, y = 0, z = 0 }
room.exits = { "east" }

room.loot = {
    {
        item = "Apple",
        amount = 1,
        drop_chance_percent = 100,
    },
    {
        item = "Potion",
        amount = 1,
        drop_chance_percent = 100,
    }
}

room.interactions = {
    touch = function (state, player, arg)
        if arg == "coffee maker" then
            world.message_player(player, "You touched the coffee maker.")
            return
        end

        world.message_player(player, string.format("There is no '%s' here that you can touch.", arg))
    end
}

return room
