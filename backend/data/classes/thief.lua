local class = {}

class.name = "Thief"
class.stats = { VIT = 8, STR = 8, AGI = 10, INT = 6, FTH = 8 }
class.scaling = { VIT = 8, STR = 8, AGI = 10, INT = 6, FTH = 8 }
class.unlocks = {
	{ level = 1, ability = "Sneak" },
}

return class
