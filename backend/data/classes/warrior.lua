local class = {}

class.name = "Warrior"
class.description = "A frontline fighter who relies on strength and draws enemies' attention."
class.stats = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.starting_items = {}
class.unlocks = {
	{ level = 2, ability = "Taunt" },
}

return class
