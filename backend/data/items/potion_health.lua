local item = {}

item.name = "Potion of Health"
item.description = "A red tonic that gives health to the drinker"
item.kind = 0 -- ITEM_KIND_CONSUMABLE
item.size = 5

item.on_use = function(user, target)
	world.log("{target} drank a health potion and regained {healing} HP.", {
		target = "Friend",
		healing = 20,
	})
end

return item
