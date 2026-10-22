local npc = {}

UNFINISHED

npc.id = "skulk"
npc.name = "Skulk"
npc.description = "A shadowy, half-visible little create with clawed hands and a large mouth. It seems interested in watching from the shadows."
npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.equipment = {}
npc.drop_table = { }

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

TO DO: turns hostile if a player shows weakness

return npc
