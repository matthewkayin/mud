local item = {}

item.name = "Dark Iron Axe"
item.description = "This axe is forged from dark iron, giving its wielder tremendous power."
item.kind = world.ItemKind.EQUIPMENT_TWO_HANDED
item.price = 300
item.size = 10

item.damage = 10
item.max_durability = 200
item.stat_bonuses = { STR = 9 }
item.stat_requirements = { STR = 12 }

return item
