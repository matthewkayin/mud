local item = {}

item.name = "Azurite Pendant"
item.description = "A silver pendant containing a supernatural blue gemstone."
item.kind = world.ItemKind.EQUIPMENT_ACCESSORY
item.price = 1000
item.size = 1

item.stat_bonuses = { INT = 3, FAI = 3 }
item.stat_requirements = { INT = 12 }

return item
