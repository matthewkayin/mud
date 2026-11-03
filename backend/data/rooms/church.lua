local room = {}

room.name = "Church"
room.description = "Has now been converted into a hardware room"
room.notes = ""

room.position = { x = 0, y = 0, z = 1 }
room.is_safe_zone = true
room.exits = {
    world.Direction.DOWN
}

room.locks = {}
room.drop_table = {}
room.chests = {}
room.fixtures = {}

return room
