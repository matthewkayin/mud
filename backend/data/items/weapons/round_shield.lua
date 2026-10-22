
local item = {}

item.name = "Round Shield"
item.prefix = "Reinforced"
item.description = "A circular shield made of wood and iron."
item.kind = world.ItemKind.EQUIPMENT_SHIELD
item.price = 50
item.size = 10

item.armor = 4
item.max_durability = 200
item.stealth_penalty = 0.25
item.stat_bonuses = {}
item.stat_requirements = { STR = 10 }

item.notes = ""

return item
