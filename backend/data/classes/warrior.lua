local class = {}

class.name = "Warrior"
class.stats = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 10, AGI = 6, INT = 6, FTH = 8 }
class.unlocks = {
	{ level = 2, ability = "Taunt" },
}

return class
