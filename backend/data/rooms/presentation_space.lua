local room = {}

room.name = "Presentation Space"
room.description =
    "You are in a wide-open room with chairs stacked against the columns. " ..
    "A makeshift, blue octopus hangs overhead."
room.notes = ""

room.position = { x = 0, y = 0, z = 0 }
room.is_safe_zone = true
room.exits = {
    world.Direction.WEST,
    world.Direction.UP
}

room.locks = {}
room.drop_table = {}
room.chests = {}

return room
