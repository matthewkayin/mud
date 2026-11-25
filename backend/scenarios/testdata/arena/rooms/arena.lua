local room = {}

room.name = "Arena"
room.description = "A sandy pit for testing combat."

room.npcs = {
    { id = "training_dummy" }
}

room.loot = {
    {
        item = "Spellbook of Heal",
        amount = 1,
        drop_chance_percent = 100
    },
    {
        item = "The Ruler of Focus",
        amount = 1,
        drop_chance_percent = 100
    }
}

return room
