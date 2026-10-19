local item = {}

item.name = "Crude Maul"
item.description = "A rudimentary implement made for chopping wood; useful for bashing."
item.kind = world.ItemKind.EQUIPMENT_TWO_HANDED
item.price = 15
item.size = 10

item.damage = 6
item.max_durability = 200
item.stat_bonuses = { STR = 1 }
item.stat_requirements = { STR = 9 }

return item
