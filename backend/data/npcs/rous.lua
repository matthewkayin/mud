local npc = {}

npc.id = "rous"
npc.name = "Giant Rat"
npc.description = "This is an unusually large, very mean rat. Some people don't think rats this size even exist."
npc.experience_worth = 15
npc.experience_worth_scaling = 2
npc.respawn_duration = world.time_minutes(2)

npc.stats = { VIT = 8, STR = 4, AGI = 8, INT = 1, FTH = 1 }
npc.scaling = { VIT = 8, STR = 4, AGI = 8, INT = 1, FTH = 1 }
npc.damage = 2
npc.armor_penetration = 0
npc.deflection = 1
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = {
    {
        item = "Leather",
        amount = { min = 1, max = 2 },
        durability_percent = 100,
        drop_chance_percent = 15
    },
    {
        item = "Monster Slime",
        amount = { min = 1, max = 2 },
        durability_percent = 100,
        drop_chance_percent = 8
    },
    {
        item = "Bezoar",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 20
    },
}

npc.infusion = {
    condition = "Disease",
    apply_chance = 0.15,
    apply_duration = world.time_minutes(60),
    uses_remaining = 2,
}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.STATIONARY

return npc
