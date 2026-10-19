local recipe = {}

recipe.name = "Potion of Health"
recipe.job = "Alchemist"
recipe.level = 1
recipe.materials = {
	{ item = "Glass Bottle", amount = 1 },
	{ item = "Bezoar", amount = 1 },
	{ item = "Fancy Plants", amount = 2 },
}
recipe.output = { item = "Potion of Health", amount = 1 }

return recipe
