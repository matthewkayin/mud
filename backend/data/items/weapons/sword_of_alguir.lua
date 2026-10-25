local item = {}

item.name = "Sword of Alguir"
item.prefix = "Sharpened"
item.description = "This powerful greatsword was forged long ago for one of the mighty knights of Alguir, before being lost to history."
item.kind = world.ItemKind.EQUIPMENT_TWO_HANDED
item.price = 1000
item.size = 10

item.damage = 12
item.armor_penetration = 0
item.max_durability = 200
item.stat_bonuses = { AGI = 5, STR = 5 }
item.stat_requirements = { STR = 12 }

item.notes = "tier 1 rare"

return item
