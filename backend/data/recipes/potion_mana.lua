local recipe = {}

recipe.name = "Potion of Mana"
recipe.job = "Alchemist"
recipe.level = 2
recipe.materials = {
	{ item = "Glass Bottle", amount = 1 },
	{ item = "Monster Horn", amount = 2 },
	{ item = "Monster Slime", amount = 2 },
}
recipe.output = { item = "Potion of Mana", amount = 1 }

return recipe
