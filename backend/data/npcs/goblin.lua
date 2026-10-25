local npc = {}

npc.id = "goblin"
npc.name = "Goblin"
npc.description = "You see a repulsive, green monster that wants to eat you."

npc.experience_worth = 100
npc.experience_worth_scaling = 25
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.damage = 0
npc.armor_penetration = 0
npc.deflection = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
