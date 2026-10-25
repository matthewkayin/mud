local item = {}

item.name = "Kitchen Knife"
item.prefix = "Sharpened"
item.description = "Originally made for cutting up ingredients and not for slaying monsters."
item.kind = world.ItemKind.EQUIPMENT_ONE_HANDED
item.price = 15
item.size = 12

item.damage = 4
item.armor_penetration = 0
item.max_durability = 50
item.stat_bonuses = { AGI = 1 }
item.stat_requirements = { STR = 2 }

item.notes = "starter gear"

return item
