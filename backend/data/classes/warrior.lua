local class = {}

class.name = "Warrior"
class.description = "Common amongst heroes in any land or time period. In dark times, strength is a cardinal virtue."
class.stats = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.starting_items = {
    { item = "Crude Maul", amount = 1 }
}
class.starting_equipment = {
    outfit = "Rags"
}
class.unlocks = {
	{ level = 2, ability = "Taunt" },
}

return class
