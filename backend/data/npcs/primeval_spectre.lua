local npc = {}

npc.id = "primeval spectre"
npc.name = "Primeval Spectre"
npc.description = "A terrifying apparation of an ancient king guards this tomb."

npc.experience_worth = 200
npc.experience_worth_scaling = 75
npc.respawn_duration = world.time_minutes(10)

npc.stats = { VIT = 8, STR = 2, AGI = 15, INT = 2, FTH = 2 }
npc.scaling = { VIT = 8, STR = 2, AGI = 15, INT = 2, FTH = 2 }
npc.damage = 10
npc.armor_penetration = 4
npc.deflection = 8
npc.damage_resistance = 7

npc.equipment = {}
npc.drop_table = {
    {
       item = "Pixie Dust",
       amount = { min = 2, max = 18 },
       durability_percent = 100,
       drop_chance_percent = 100
   },
   {
       item = "Monster Slime",
       amount = { min = 1, max = 3 },
       durability_percent = 100,
       drop_chance_percent = 100
   },
   {
       item = "Gold",
       amount = { min = 50, max = 300 },
       durability_percent = 100,
       drop_chance_percent = 100
   },
   {
       item = "Sword of Alguir",
       amount = 1,
       durability_percent = 100,
       drop_chance_percent = 10
   },
   {
       item = "Greenbolt",
       amount = 1,
       durability_percent = 100,
       drop_chance_percent = 10
    },
    {
        item = "Spellbook of Chain Lightning",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 12
    },
    {
        item = "Dark Iron Dirk",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 4
    },
    {
        item = "Dark Iron Axe",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 4
    },
    {
        item = "Dark Iron Longsword",
        amount = 1,
        durability_percent = 100,
        drop_chance_percent = 4
    },

}

npc.starting_disposition = world.NpcDisposition.HOSTILE
npc.movement_type = world.NpcMovementType.STATIONARY

return npc
