local item = {}

item.name = "Dark Iron Longsword"
item.prefix = "Sharpened"
item.description = "This longsword is made from dark iron, and gives its wielder dreadful strength and speed."
item.kind = world.ItemKind.EQUIPMENT_ONE_HANDED
item.price = 300
item.size = 10

item.damage = 8
item.armor_penetration = 4
item.max_durability = 150
item.stat_bonuses = { AGI = 1, STR = 1 }
item.stat_requirements = { STR = 11 }

item.notes = "crafted only"

return item
