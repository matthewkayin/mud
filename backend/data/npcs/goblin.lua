local npc = {}

npc.id = "goblin"
npc.name = "Goblin"
npc.description = "You see a repulsive, green monster that wants to eat you."

npc.experience_worth = 20
npc.experience_worth_scaling = 5
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 8, STR = 4, AGI = 4, INT = 2, FTH = 4 }
npc.scaling = { VIT = 8, STR = 4, AGI = 4, INT = 2, FTH = 4 }
npc.damage = 6
npc.armor_penetration = 1
npc.deflection = 6

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
