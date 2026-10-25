local npc = {}

---UNFINISHED

npc.id = "hill giant"
npc.name = "Hill Giant"
npc.description = "This Hill Giant is known to the local townspeople, who regard it with fear."

npc.experience_worth = 100
npc.experience_worth_scaling = 25
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.damage = 0
npc.armor_penetration = 0
npc.deflection = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.STATIONARY

---TO DO: make him wander around in neutral disposition until he finds a happy cow, which he then eats, and then naps
---too long without a cow snack and he gets angry, becoming hostile.
return npc
