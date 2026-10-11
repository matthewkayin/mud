local item = {}

item.name = "Sword"
item.description = "A pointy metal stick with a handle."
item.kind = world.ItemKind.EQUIPMENT_ONE_HANDED
item.size = 10

item.damage = 5
item.max_durability = 100
item.stat_bonuses = { STR = 2 }
item.stat_requirements = {}

return item
