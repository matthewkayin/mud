local class = {}

class.name = "Priest"
class.description = "Some priests have no flock, and instead travel the land to slay evil and assist heroes. Can use powerful support magicks."
class.stats = { VIT = 6, STR = 6, AGI = 8, INT = 8, FTH = 10 }
class.scaling = { VIT = 6, STR = 6, AGI = 8, INT = 8, FTH = 10 }
class.starting_items = {
    { item = "Spellbook of Cure", amount = 1 },
    { item = "Walking Stick", amount = 1 },
}
class.starting_equipment = {
    outfit = "Rags"
}
class.unlocks = {
    { level = 1, spell = "Blessing of Might" },
    { level = 2, spell = "Purify" }
}

return class
