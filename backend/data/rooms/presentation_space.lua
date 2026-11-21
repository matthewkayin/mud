local room = {}

room.name = "Presentation Space"
room.description = "There is a blue octopus in this room"

room.npcs = {
    { id = "happy_cow" },
    { id = "happy_cow" }
}

room.loot = {
    {
        item = "Leather",
        amount = 3,
        drop_chance_percent = 100,
    }
}

return room
