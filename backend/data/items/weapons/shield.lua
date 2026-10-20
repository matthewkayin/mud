
local item = {}

item.name = "Shield"
item.description = "It's just a fucking shield, okay? Do you not know what that is?"
item.kind = world.ItemKind.EQUIPMENT_SHIELD
item.price = 50
item.size = 10

item.armor = 4
item.max_durability = 200
item.stealth_penalty = 0.25
item.stat_bonuses = {}
item.stat_requirements = { STR = 10 }

return item
