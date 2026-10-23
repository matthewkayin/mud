local npc = {}

---UNFINISHED

npc.id = "goblin apple"
npc.name = "Goblin"
npc.description = "You see a repulsive, green monster that wants to eat you."

npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

---This guy tries to get fruit from a nearby tree, and when he gets it and eats it he is satisfied for a while.

return npc
