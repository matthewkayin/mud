local room = {}

room.name = "Arena"
room.description = "A sandy pit for testing combat."

room.npcs = {
    { id = "training_dummy" }
}

room.loot = {
    { item = "Leather Armor", amount = 1, drop_chance_percent = 100 },
}

return room
