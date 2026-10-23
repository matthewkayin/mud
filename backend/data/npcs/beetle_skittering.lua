local npc = {}

npc.id = "skittering beetle"
npc.name = "Skittering Beetle"
npc.description = "This scuttling little villain has a hard little shell and sharp little pincers."
npc.experience_worth = 50
npc.experience_worth_scaling = 15

npc.stats = { VIT = 1, STR = 2, AGI = 4, INT = 1, FTH = 1 }
npc.scaling = { VIT = 1, STR = 2, AGI = 4, INT = 1, FTH = 1 }
npc.bonus_damage = 1
npc.bonus_armor = 4
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.STATIONARY

return npc
