local recipe = {}

recipe.name = "Repair Kit"
recipe.job = "Blacksmith"
recipe.level = 1
recipe.materials = {
    { item = "Common Lumber", amount = 1 },
    { item = "Leather", amount = 1 },
	{ item = "Iron", amount = 1 },
}
recipe.output = { item = "Repair Kit", amount = 5 }

return recipe
