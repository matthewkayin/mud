local npc = {}

npc.id = "spider"
npc.name = "Spider"
npc.description = "A giant, hairy spider with scutters about, glaring at you with its lamp-like eyes."

npc.experience_worth = 40
npc.experience_worth_scaling = 5
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 12, STR = 2, AGI = 10, INT = 2, FTH = 4 }
npc.scaling = { VIT = 9, STR = 2, AGI = 10, INT = 2, FTH = 4 }
npc.damage = 3
npc.armor_penetration = 1
npc.deflection = 6
npc.damage_resistance = 1

npc.equipment = {
    {
        item = "Monster Slime",
        amount = { min = 1, max = 2 },
        durability_percent = 100,
        drop_chance_percent = 20
    },
    {
        item = "Pixie Dust",
        amount = { min = 1, max = 3 },
        durability_percent = 100,
        drop_chance_percent = 10
    },
    {
        item = "Venom Gland",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 80
    },
}

npc.infusion = {
    condition = "Poison",
    apply_chance = 0.25,
    apply_duration = world.time_minutes(5),
    uses_remaining = 2,
}

npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
