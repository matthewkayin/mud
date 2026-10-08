local item = {}

item.name = "Potion of Mana"
item.description = "A blue tonic that gives mana to the drinker."
item.kind = world.ItemKind.CONSUMABLE
item.size = 5

item.on_use = function(user, target)
	world.log("{target} drank a mana potion and regained {mana} MP.", {
		target = "Friend",
		mana = 20,
	})
end

return item
