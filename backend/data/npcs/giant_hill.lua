local npc = {}

---UNFINISHED

npc.id = "hill giant"
npc.name = "Hill Giant"
npc.description = "This Hill Giant is known to the local townspeople, who regard it with fear."

npc.experience_worth = 150
npc.experience_worth_scaling = 75
npc.respawn_duration = world.time_minutes(10)

npc.stats = { VIT = 20, STR = 10, AGI = 3, INT = 2, FTH = 2 }
npc.scaling = { VIT = 20, STR = 10, AGI = 3, INT = 2, FTH = 2 }
npc.damage = 10
npc.armor_penetration = 10
npc.deflection = 6
npc.damage_resistance = 4

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.WANDER

---TO DO: make him wander around in neutral disposition until he finds a happy cow, which he then eats, and then naps
---too long without a cow snack and he gets angry, becoming hostile.
return npc
