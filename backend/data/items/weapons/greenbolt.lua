local item = {}

item.name = "Greenbolt"
item.prefix = "Sharpened"
item.description = "This green and gold Shortsword was crafted by anicent Elves as a gift for a Hobbit."
item.kind = world.ItemKind.EQUIPMENT_ONE_HANDED
item.price = 1000
item.size = 5

item.damage = 9
item.armor_penetration = 0
item.max_durability = 150
item.stat_bonuses = { AGI = 12 }
item.stat_requirements = { STR = 11 }

item.notes = "tier 1 rare"

return item
