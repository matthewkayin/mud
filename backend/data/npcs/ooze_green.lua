local npc = {}

npc.id = "green ooze"
npc.name = "Green Ooze"
npc.description = "You see a bizarre, green mass of goop that wants absorb you. Does cutting this thing even hurt it?"

npc.experience_worth = 100
npc.experience_worth_scaling = 25
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 2, STR = 6, AGI = 1, INT = 1, FTH = 8 }
npc.scaling = { VIT = 2, STR = 6, AGI = 1, INT = 1, FTH = 8 }
npc.damage = 0
npc.armor_penetration = 0
npc.deflection = 9

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.STATIONARY

return npc
