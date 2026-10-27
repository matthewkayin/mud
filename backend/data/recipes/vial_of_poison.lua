local recipe = {}

recipe.name = "Vial of Poison"
recipe.job = "Alchemist"
recipe.level = 1
recipe.materials = {
	{ item = "Glass Bottle", amount = 1 },
	{ item = "Venom Gland", amount = 1 },
}
recipe.output = { item = "Vial of Poison", amount = 1 }

return recipe
