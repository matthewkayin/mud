local npc = {}

npc.id = "goblin"
npc.name = "Goblin"
npc.description = "You see a repulsive, green monster that wants to eat you."

npc.experience_worth = 20
npc.experience_worth_scaling = 2
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 8, STR = 1, AGI = 4, INT = 2, FTH = 4 }
npc.scaling = { VIT = 8, STR = 1, AGI = 4, INT = 2, FTH = 4 }
npc.damage = 3
npc.armor_penetration = 0
npc.deflection = 4
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = {
    {
        item = "Apple",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 2
    },
    {
        item = "Gold",
        amount = { min = 1, max = 12 },
        durability_percent = 100,
        drop_chance_percent = 25
    },
    {
        item = "Iron",
        amount = { min = 1, max = 2 },
        durability_percent = 100,
        drop_chance_percent = 3
    },
    {
        item = "Monster Horn",
        amount = { min = 1, max = 2 },
        durability_percent = 100,
        drop_chance_percent = 8
    },
    {
        item = "Bezoar",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 5
    },
    {
        item = "Crude Iron Plating",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 8
    },
    {
        item = "Iron Dagger",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 6
    },
    {
        item = "Iron Axe",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 6
    },
    {
        item = "Round Shield",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 8
    },
}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
