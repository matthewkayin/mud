local npc = {}

npc.id = "goblin"
npc.name = "Goblin"
npc.description = "You see a repulsive, green monster that wants to eat you."

npc.experience_worth = 10
npc.experience_worth_scaling = 2
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 9, STR = 4, AGI = 4, INT = 2, FTH = 4 }
npc.scaling = { VIT = 9, STR = 4, AGI = 4, INT = 2, FTH = 4 }
npc.damage = 4
npc.armor_penetration = 0
npc.deflection = 4
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

return npc
