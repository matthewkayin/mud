local recipe = {}

recipe.name = "Antidote"
recipe.job = "Alchemist"
recipe.level = 1
recipe.materials = {
	{ item = "Glass Bottle", amount = 1 },
	{ item = "Fancy Plants", amount = 2 },
	{ item = "Wild Mushrooms", amount = 2 },
}
recipe.output = { item = "Antidote", amount = 1 }

return recipe
