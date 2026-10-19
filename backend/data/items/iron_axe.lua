local item = {}

item.name = "Iron Axe"
item.description = "A heavy, powerful axe for splitting monsters in two."
item.kind = world.ItemKind.EQUIPMENT_TWO_HANDED
item.price = 100
item.size = 10

item.damage = 7
item.max_durability = 150
item.stat_bonuses = { STR = 5 }
item.stat_requirements = { STR = 10 }

return item
