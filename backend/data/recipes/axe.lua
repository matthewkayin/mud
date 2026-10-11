local recipe = {}

recipe.name = "Axe"
recipe.job = "Blacksmith"
recipe.level = 1
recipe.materials = {
	{ item = "Sword", amount = 2 },
	{ item = "Dummy Material", amount = 5 },
}
recipe.output = { item = "Axe", amount = 1 }

return recipe
