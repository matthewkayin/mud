local recipe = {}

recipe.name = "Weapon Reinforcement Kit"
recipe.job = "Blacksmith"
recipe.level = 2
recipe.materials = {
    { item = "Common Lumber", amount = 1 },
    { item = "Leather", amount = 2 },
	{ item = "Iron", amount = 2 },
}
recipe.output = { item = "Repair Kit", amount = 1 }

return recipe
