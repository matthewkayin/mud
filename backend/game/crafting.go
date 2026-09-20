package game

import (
	"fmt"
)

type Recipe int
const (
	RECIPE_HEALTH_POTION = iota
	RECIPE_MANA_POTION
	RECIPE_CURE_BOOK
	RECIPE_FIREBOLT_BOOK
	RECIPE_SWORD
	RECIPE_AXE
)

type Material struct {
	id ItemId
	amount int32
}

type RecipeData struct {
	name string
	job CharacterJob
	level int32
	materials []Material
	output Item
}


var RECIPE_DATA = map[Recipe]*RecipeData {

	RECIPE_HEALTH_POTION: {
		name: "Potion of Health",
		job: CHARACTER_JOB_ALCHEMIST,
		level: 1,
		materials: []Material {
			{id: ITEM_DUMMY_MATERIAL, amount: 10},
		},
		output: Item {
			Id: ITEM_POTION_HEALTH,
			Amount: 1,
		},
	},

	RECIPE_MANA_POTION: {
		name: "Potion of Mana",
		job: CHARACTER_JOB_ALCHEMIST,
		level: 2,
		materials: []Material {
			{id: ITEM_DUMMY_MATERIAL, amount: 10},
		},
		output: Item {
			Id: ITEM_POTION_MANA,
			Amount: 1,
		},
	},

	RECIPE_SWORD: {
		name: "Sword",
		job: CHARACTER_JOB_BLACKSMITH,
		level: 1,
		materials: []Material {
			{id: ITEM_AXE, amount: 2},
			{id: ITEM_DUMMY_MATERIAL, amount: 5},
		},
		output: Item {
			Id: ITEM_SWORD,
			Amount: 1,
		},
	},

	RECIPE_AXE: {
		name: "Axe",
		job: CHARACTER_JOB_BLACKSMITH,
		level: 1,
		materials: []Material {
			{id: ITEM_SWORD, amount: 2},
			{id: ITEM_DUMMY_MATERIAL, amount: 5},
		},
		output: Item {
			Id: ITEM_AXE,
			Amount: 1,
		},
	},
}

//add recipe to character recipes list
func (recipe Recipe) LearnRecipe (player *Player) {
	recipeData := RECIPE_DATA[recipe]
	player.character.RecipesKnown = append(player.character.RecipesKnown, recipe)
	*player.inbox <- fmt.Sprintf("You have learned the recipe for: %s.", recipeData.name)
}

//craft an item from a recipe
func (recipe Recipe) Craft (inventory *Inventory) (bool) {
	recipeData := RECIPE_DATA[recipe]

	//first check that the materials are there
	for _, ingredient := range recipeData.materials {
		amountOfIngredient := inventory.AmountOf(ingredient.id)
		if amountOfIngredient < ingredient.amount {
			return false
		}
	}

	//check every item in inventory against the grocery list
	for _, ingredient := range recipeData.materials {
		amountToRemove := ingredient.amount
		for amountToRemove > 0 {
    		index, _ := inventory.FindItem(ingredient.id)
      		item := inventory.RemoveItems(index, amountToRemove)
       		amountToRemove -= item.Amount
		}
	}

	//and add item
	inventory.AddItem(recipeData.output)
	return true
}
