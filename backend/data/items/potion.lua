local item = {}

item.name = "Potion"
item.description = "This is a potion"
item.kind = world.ItemKind.CONSUMABLE
item.price = 20
item.size = 1
item.target_type = world.TargetType.NONE

function item.on_use(self)
	world.message_player(self, "You drink the potion. It tastes faintly of cherries.")
end

return item
