local item = {}

item.name = "Test Shield"
item.description = "A shield."
item.kind = world.ItemKind.SHIELD
item.price = 5
item.size = 1

item.max_durability = 100
item.reinforced_prefix = "badass"

item.stats = {
    [world.Stat.DEFLECTION] = 6,
    [world.Stat.DEFLECTION_AMOUNT] = 4,
}

return item
