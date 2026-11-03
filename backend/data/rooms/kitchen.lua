local room = {}

room.name = "Kitchen"
room.description = "This is the kitchen"
room.notes = ""

room.position = { x = -1, y = 0, z = 0 }
room.is_safe_zone = true
room.exits = {
    world.Direction.EAST
}

room.locks = {}
room.drop_table = {}
room.chests = {}

return room
