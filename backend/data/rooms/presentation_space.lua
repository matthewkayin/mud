local room = {}

room.name = "Presentation Space"
room.description = "There is a blue octopus in this room"
room.position = { x = 0, y = 0, z = 0 }
room.exits = { "west", "up" }

room.npcs = {
    { id = "happy_cow" }
}

return room
