local item = {}

item.name = "Test Shield"
item.description = "A shield."
item.kind = world.ItemKind.SHIELD
item.price = 5
item.size = 1

item.max_durability = 100
item.prefix = "badass"

item.required_rank = 0
item.deflection_chance = 6
item.deflection_bonus = 4
item.damage_resistance = 0

item.evade_modifier = 0
item.sneak_modifier = 0
item.initiative_modifier = 0

return item
