local recipe = {}

recipe.name = "Iron Axe"
recipe.job = "Blacksmith"
recipe.level = 1
recipe.materials = {
    { item = "Leather", amount = 1 },
    { item = "Common Lumber", amount = 1 },
	{ item = "Iron", amount = 3 },
}
recipe.output = { item = "Iron Axe", amount = 1 }

return recipe
