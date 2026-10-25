local npc = {}

npc.id = "manticore"
npc.name = "Manticore"
npc.description = "A gargantuan lion with the face of a man and the tail of a scoprion."

npc.experience_worth = 200
npc.experience_worth_scaling = 25
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.damage = 6
npc.armor_penetration = 0
npc.deflection = 4

npc.equipment = {}

npc.infusion = {
    condition = "Poison",
    apply_chance = 0.5,
    apply_duration = world.time_minutes(5),
    uses_remaining = 10,
}

npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
