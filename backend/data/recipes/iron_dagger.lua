local recipe = {}

recipe.name = "Iron Dagger"
recipe.job = "Blacksmith"
recipe.level = 1
recipe.materials = {
    { item = "Leather", amount = 1 },
	{ item = "Iron", amount = 2 },
}
recipe.output = { item = "Iron Dagger", amount = 1 }

return recipe
