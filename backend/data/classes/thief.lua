local class = {}

class.name = "Thief"
class.description = "Thieves prefer to sneak and steal over direct conflict with monsters. Can scout and unlock doors for a party."
class.stats = { VIT = 8, STR = 8, AGI = 10, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 8, AGI = 10, INT = 6, FTH = 8 }
class.starting_items = {
    { item = "Kitchen Knife", amount = 1 },
}
class.starting_equipment = {
    outfit = "Rags"
}

class.unlocks = {
	{ level = 1, ability = "Sneak" },
	{ level = 2, ability = "Unlock" },
}

return class
