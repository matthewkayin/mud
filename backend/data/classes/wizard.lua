local class = {}

class.name = "Wizard"
class.description = "Wizards are masters of magic! They have deep and broad mastery over nearly all schools of magic."
class.stats = { VIT = 6, STR = 6, AGI = 8, INT = 10, FTH = 8 }
class.scaling = { VIT = 6, STR = 6, AGI = 8, INT = 10, FTH = 8 }
class.starting_items = {
    { item = "Spellbook of Firebolt", amount = 1 },
    { item = "Walking Stick", amount = 1 },
}
class.starting_equipment = {
    outfit = "Rags"
}
class.unlocks = {}

return class
