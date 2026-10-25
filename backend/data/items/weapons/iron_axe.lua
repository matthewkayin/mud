local item = {}

item.name = "Iron Axe"
item.prefix = "Sharpened"
item.description = "A heavy, powerful axe for splitting monsters in two."
item.kind = world.ItemKind.EQUIPMENT_TWO_HANDED
item.price = 100
item.size = 10

item.damage = 7
item.armor_penetration = 3
item.max_durability = 150
item.stat_bonuses = {}
item.stat_requirements = { STR = 10 }

item.notes = ""

return item
