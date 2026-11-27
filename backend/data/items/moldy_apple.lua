local item = {}

item.name = "Moldy Apple"
item.description = "This apple looks kinda sus"
item.kind = world.ItemKind.CONSUMABLE
item.price = 2
item.size = 1
item.target_type = world.TargetType.NONE

function item.on_use(self)
    world.mob_add_condition(self, "Poison", world.time_seconds(30))
end

return item
