local npc = {}

npc.id = "rous"
npc.name = "Giant Rat"
npc.description = "This is an unusually large, very mean rat. Some people don't think rats this size even exist."
npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 3, STR = 4, AGI = 4, INT = 1, FTH = 1 }
npc.scaling = { VIT = 3, STR = 4, AGI = 4, INT = 1, FTH = 1 }
npc.bonus_damage = 1
npc.bonus_armor = 1
npc.equipment = {}
npc.drop_table = {}

npc.infusion = {
    condition = "Disease",
    apply_chance = 0.1,
    apply_duration = world.time_minutes(60),
    uses_remaining = 2,
}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
