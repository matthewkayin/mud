local item = {}

item.name = "Apple"
item.description = "And this is an apple"
item.kind = world.ItemKind.CONSUMABLE
item.price = 2
item.size = 1
item.target_type = world.TargetType.NONE

function item.on_use(self)
	world.message_player(self, "You eat the apple. Crunchy.")
end

return item
