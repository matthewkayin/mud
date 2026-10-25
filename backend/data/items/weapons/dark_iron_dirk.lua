local item = {}

item.name = "Dark Iron Dirk"
item.prefix = "Sharpened"
item.description = "This short blade is made from dark iron, making it exceedingly deadly."
item.kind = world.ItemKind.EQUIPMENT_ONE_HANDED
item.price = 250
item.size = 5

item.damage = 7
item.armor_penetration = 0
item.max_durability = 150
item.stat_bonuses = { AGI = 9 }
item.stat_requirements = { STR = 6 }

item.notes = "crafter only"

return item
