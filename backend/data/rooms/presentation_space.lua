local room = {}

room.name = "Presentation Space"
room.description = "There is a blue octopus in this room"

room.npcs = {
    { id = "happy_cow" },
    { id = "small_cow" },
    { id = "brown_cow" }
}

room.loot = {
    {
        item = "Spellbook of Firebolt",
        amount = 1,
        drop_chance_percent = 100,
    },
    {
        item = "The Ruler of Focus",
        amount = 1,
        drop_chance_percent = 100,
    }
}

return room
