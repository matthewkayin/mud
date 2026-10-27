local npc = {}

npc.id = "green ooze"
npc.name = "Green Ooze"
npc.description = "You see a bizarre, green mass of goop that wants absorb you. Does cutting this thing even hurt it?"

npc.experience_worth = 40
npc.experience_worth_scaling = 25
npc.respawn_duration = world.time_minutes(2)

npc.stats = { VIT = 4, STR = 6, AGI = 1, INT = 1, FTH = 8 }
npc.scaling = { VIT = 4, STR = 6, AGI = 1, INT = 1, FTH = 8 }
npc.damage = 6
npc.armor_penetration = 2
npc.deflection = 0
npc.damage_resistance = 9

npc.equipment = {}
npc.drop_table = {

    {
        item = "Gold",
        amount = { min = 1, max = 1000 },
        durability_percent = 100,
        drop_chance_percent = 1
    },
    {
        item = "Iron",
        amount = { min = 1, max = 2 },
        durability_percent = 100,
        drop_chance_percent = 3
    },
    {
        item = "Monster Slime",
        amount = { min = 1, max = 8 },
        durability_percent = 100,
        drop_chance_percent = 100
    },
    {
        item = "Crude Iron Plating",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 2
    },
    {
        item = "Iron Dagger",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 2
    },
    {
        item = "Iron Axe",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 2
    },
    {
        item = "Iron Longsword",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 2
    },
    {
        item = "Round Shield",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 2
    },
    {
        item = "Rags",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 15
    },
}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.STATIONARY

return npc
