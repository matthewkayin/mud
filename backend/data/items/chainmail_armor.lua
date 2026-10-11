local item = {}

item.name = "Chainmail Armor"
item.description = "Armor made of interlocking link of metal"
item.kind = world.ItemKind.EQUIPMENT_OUTFIT
item.size = 20

item.armor = 5
item.max_durability = 200
item.stealth_penalty = 0.25
item.stat_bonuses = {}
item.stat_requirements = { STR = 10 }

return item
