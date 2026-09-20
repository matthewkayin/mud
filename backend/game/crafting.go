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
	Id ItemId
	Amount int32
}

type RecipeData struct {
	name string
	job CharacterJob
	level int
	materials []Material
	output Item
}


var RECIPE_DATA = map[Recipe]*RecipeData {

	RECIPE_HEALTH_POTION: {
		name: "Potion of Health",
		job: CHARACTER_JOB_ALCHEMIST,
		level: 1,
		materials: []Material {
			{Id: ITEM_DUMMY_MATERIAL, Amount: 10},
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
			{Id: ITEM_DUMMY_MATERIAL, Amount: 10},
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
			{Id: ITEM_AXE, Amount: 2},
			{Id: ITEM_DUMMY_MATERIAL, Amount: 5},
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
			{Id: ITEM_SWORD, Amount: 2},
			{Id: ITEM_DUMMY_MATERIAL, Amount: 5},
		},
		output: Item {
			Id: ITEM_SWORD,
			Amount: 1,
		},
	},
}

//check for legality of recipe and then add to character recipes list
func (recipe Recipe) LearnRecipe (character *Character) (string, bool) {
	recipeData := RECIPE_DATA[recipe]

	if recipeData.job != character.Job {
		return fmt.Sprintf("You must be a %s to learn that recipe.", JOB_DATA[recipeData.job].Name), false
	}

	if recipeData.level > int(character.Data.Level) {
		return "Your level is not high enough to learn this recipe.", false
	}

	character.RecipesKnown = append(character.RecipesKnown, recipe)
	return fmt.Sprintf("You have learned the recipe for: %s.", ITEM_DATA[recipeData.output.Id].name), true
}

//craft an item from a recipe
func (recipe Recipe) Craft (inventory *Inventory) (bool) {
recipeData := RECIPE_DATA[recipe]

	//first check that the materials are there
	for _, ingredient := range recipeData.materials {
		_, hasIngredients := inventory.CheckForItem(ingredient.Id, ingredient.Amount)
		if !hasIngredients {
			return false
		}
	}

	//then remove the ingredients
	var groceryList []Material = recipeData.materials
	//check every item in inventory against the grocery list
	for index, item := range inventory.Items {
		for _, ingredient := range groceryList {
			if item.Id == ingredient.Id && ingredient.Amount > 0 {
				//remove the item from the inventory while also clearing it from the grocery list
				removed := min(item.Amount, ingredient.Amount)
				inventory.Items[index].Amount = max(item.Amount - ingredient.Amount, 0)
				ingredient.Amount -= removed
				if inventory.Items[index].Amount == 0 {
					inventory.RemoveItem(index)
				}
				continue
			}
		}
	}

	//and add item
	inventory.AddItem(recipeData.output)
	return true
}

//TO DO: SPELL BOOK CRAFTING BUT WITH ITS RECIPE BEING YOU KNOWLEDGE OF THE SPELL!!
