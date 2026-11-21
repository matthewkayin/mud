local item = {}

item.name = "Test Armor"
item.description = "Some armor, and it's made out of leather!"
item.kind = world.ItemKind.LIGHT_ARMOR
item.price = 5
item.size = 1

item.max_durability = 100
item.reinforced_prefix = "badass"

item.deflection = 6
item.deflection_bonus = 4
item.resistance = 0

item.stats = {
    [world.Stat.DEFLECTION] = 6,
    [world.Stat.DEFLECTION_AMOUNT] = 4,
}

return item
