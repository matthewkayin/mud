local item = {}

item.name = "Bread"
item.description = "A crusty loaf of bread"
item.kind = world.ItemKind.CONSUMABLE
item.price = 2
item.size = 1
item.target_type = world.TargetType.NONE

function item.on_use(self)
	world.message_player(self, "You eat the bread. Filling.")
end

return item
