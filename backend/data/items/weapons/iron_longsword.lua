local item = {}

item.name = "Iron Longsword"
item.description = "You really don't know what a sword is?"
item.kind = world.ItemKind.EQUIPMENT_ONE_HANDED
item.price = 150
item.size = 10

item.damage = 6
item.max_durability = 150
item.stat_bonuses = { AGI = 2, STR = 2 }
item.stat_requirements = { STR = 9 }

item.notes = ""

return item
