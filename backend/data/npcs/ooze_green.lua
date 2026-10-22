local npc = {}

UNFINISHED

npc.id = "goblin"
npc.name = "Goblin"
npc.description = "You see a repulsive, green monster that wants to eat you."

npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.WANDER

TO DO: weapons don't work properly, take durability loss, maybe can get sucked into its equipment and dropped on kill?
and then also this badboy gets insta-zapped by fire or lightning

return npc
