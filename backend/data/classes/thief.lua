local class = {}

class.name = "Thief"
class.description = "A nimble rogue who strikes from the shadows and slips away unseen."
class.stats = { VIT = 8, STR = 8, AGI = 10, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 8, AGI = 10, INT = 6, FTH = 8 }
class.starting_items = {
    { item = "Kitchen Knife", amount = 1 },
}
class.unlocks = {
	{ level = 1, ability = "Sneak" },
	{ level = 2, ability = "Unlock" },
}

return class
