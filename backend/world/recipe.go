package world

import (
	"math"
)

const RECIPE_OUTPUT_MAX_DURABILITY = math.MaxInt32

type Recipe int
const (
	RECIPE_HEALTH_POTION = iota
	RECIPE_MANA_POTION
	RECIPE_CURE_BOOK
	RECIPE_FIREBOLT_BOOK
	RECIPE_SWORD
	RECIPE_AXE
)

type RecipeMaterial struct {
	Id ItemId
	Amount int32
}

type RecipeData struct {
	Name string
	Job JobId
	Level int32
	Materials []RecipeMaterial
	Output Item
}

var RECIPE_DATA = map[Recipe]*RecipeData {
	RECIPE_HEALTH_POTION: {
		Name: "Potion of Health",
		Job: JOB_ALCHEMIST,
		Level: 1,
		Materials: []RecipeMaterial {
			{ Id: ITEM_DUMMY_MATERIAL, Amount: 10 },
		},
		Output: Item {
			Id: ITEM_POTION_HEALTH,
			Amount: 1,
		},
	},

	RECIPE_MANA_POTION: {
		Name: "Potion of Mana",
		Job: JOB_ALCHEMIST,
		Level: 2,
		Materials: []RecipeMaterial {
			{ Id: ITEM_DUMMY_MATERIAL, Amount: 10 },
		},
		Output: Item {
			Id: ITEM_POTION_MANA,
			Amount: 1,
		},
	},

	RECIPE_SWORD: {
		Name: "Sword",
		Job: JOB_BLACKSMITH,
		Level: 1,
		Materials: []RecipeMaterial {
			{ Id: ITEM_AXE, Amount: 2 },
			{ Id: ITEM_DUMMY_MATERIAL, Amount: 5 },
		},
		Output: Item {
			Id: ITEM_SWORD,
			Amount: 1,
			Durability: RECIPE_OUTPUT_MAX_DURABILITY,
		},
	},

	RECIPE_AXE: {
		Name: "Axe",
		Job: JOB_BLACKSMITH,
		Level: 1,
		Materials: []RecipeMaterial {
			{ Id: ITEM_SWORD, Amount: 2 },
			{ Id: ITEM_DUMMY_MATERIAL, Amount: 5 },
		},
		Output: Item {
			Id: ITEM_AXE,
			Amount: 1,
			Durability: RECIPE_OUTPUT_MAX_DURABILITY,
		},
	},
}

func (recipeData *RecipeData) CreateOutput() Item {
	output := recipeData.Output
	if output.Durability == RECIPE_OUTPUT_MAX_DURABILITY {
		output.Durability = output.GetMaxDurability()
	}

	return output
}
