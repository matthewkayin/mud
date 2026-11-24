local room = {}

room.name = "Arena"
room.description = "A sandy pit for testing equipment durability."

room.npcs = {
    { id = "training_dummy" }
}

-- Each chest holds an Iron Sword, a Wooden Shield and Leather Armor. The items have a max durability of 100, so
-- durability_percent is also each item's durability
local function chest(name, durabilityPercent)
    return {
        name = name,
        description = "A chest of equipment for testing durability.",
        loot = {
            { item = "Iron Sword", amount = 1, drop_chance_percent = 100, durability_percent = durabilityPercent },
            { item = "Wooden Shield", amount = 1, drop_chance_percent = 100, durability_percent = durabilityPercent },
            { item = "Leather Armor", amount = 1, drop_chance_percent = 100, durability_percent = durabilityPercent },
        },
    }
end

room.chests = {
    -- Damaged, one hit from broken
    chest("Cracked Chest", 1),
    -- Damaged
    chest("Worn Chest", 25),
    -- Half durability, one hit from damaged (an item is damaged below half its max durability)
    chest("Scuffed Chest", 50),
    -- Full durability
    chest("Plain Chest", 100),
    -- Reinforced, one hit from losing its reinforcement
    chest("Honed Chest", 101),
    -- Fully reinforced
    chest("Masterwork Chest", 200),
}

return room
