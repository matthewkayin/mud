local item = {}

item.name = "Chainmail Armor"
item.prefix = "Reinforced"
item.description = "Armor made of interlocking links of metal."
item.kind = world.ItemKind.EQUIPMENT_OUTFIT
item.price = 100
item.size = 20

item.deflection = 9
item.damage_resistance = 1
item.max_durability = 200
item.stealth_penalty = 0.25
item.stat_bonuses = {}
item.stat_requirements = { STR = 10 }

item.notes = "tier 1 common"

return item
