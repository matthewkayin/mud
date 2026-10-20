local class = {}

class.name = "Priest"
class.description = "A servant of the divine who blesses allies and mends their wounds."
class.stats = { VIT = 6, STR = 6, AGI = 8, INT = 8, FTH = 10 }
class.scaling = { VIT = 6, STR = 6, AGI = 8, INT = 8, FTH = 10 }
class.starting_items = {
    { amount = 1, item = "Spellbook of Cure"},
    { amount = 1, item = "Walking Stick" },
    { amount = 20, item = "Repair Kit" },
}
class.starting_equipment = {
    outfit = "Rags"
}
class.unlocks = {
    { level = 1, spell = "Blessing of Might" },
    { level = 2, spell = "Purify" }
}

return class
