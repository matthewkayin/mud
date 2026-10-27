local item = {}

item.name = "Crude Iron Plating"
item.prefix = "Reinforced"
item.description = "Two rusted iron slabs used to defend your torso."
item.kind = world.ItemKind.EQUIPMENT_OUTFIT
item.price = 300
item.size = 25

item.deflection = 7
item.damage_resistance = 0
item.max_durability = 200
item.stealth_penalty = 0.25
item.stat_bonuses = {}
item.stat_requirements = { STR = 11 }

item.notes = "tier 1 common"

return item
