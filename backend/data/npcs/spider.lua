local npc = {}

npc.id = "spider"
npc.name = "Spider"
npc.description = "A giant, hairy spider with scutters about, glaring at you with its lamp-like eyes."

npc.experience_worth = 100
npc.experience_worth_scaling = 25
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.damage = 2
npc.armor_penetration = 2
npc.deflection = 2

npc.equipment = {}

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
