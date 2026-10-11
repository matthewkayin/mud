local recipe = {}

recipe.name = "Sword"
recipe.job = "Blacksmith"
recipe.level = 1
recipe.materials = {
	{ item = "Axe", amount = 2 },
	{ item = "Dummy Material", amount = 5 },
}
recipe.output = { item = "Sword", amount = 1 }

return recipe
