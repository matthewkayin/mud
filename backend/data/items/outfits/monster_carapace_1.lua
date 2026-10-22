local item = {}

item.name = "Carapace"
item.prefix = "Hardened"
item.description = "The carapace of a monster."
item.kind = world.ItemKind.EQUIPMENT_OUTFIT
item.price = 1
item.size = 1

item.armor = 8
item.max_durability = 1000
item.stealth_penalty = 0.25
item.stat_bonuses = {}
item.stat_requirements = {}

item.notes = "NPC equipment only"

return item
