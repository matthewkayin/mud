local class = {}

class.name = "Wizard"
class.description = "A scholar of the arcane who hurls destructive spells from afar."
class.stats = { VIT = 6, STR = 6, AGI = 8, INT = 10, FTH = 8 }
class.scaling = { VIT = 6, STR = 6, AGI = 8, INT = 10, FTH = 8 }
class.starting_items = {
    { item = "Spellbook of Firebolt", amount = 1 },
    { item = "Walking Stick", amount = 1 },
}
class.unlocks = {}

return class
