package world

const RECIPE_OUTPUT_MAX_DURABILITY = -1

type RecipeId int

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

var RECIPE_DATA []*RecipeData

func (recipeData *RecipeData) NetItemSize(batchAmount int32) int32 {
	var netSize int32 = 0

	// Subtract from net size for each material
	for index := range len(recipeData.Materials) {
		material := &recipeData.Materials[index]
		netSize -= ITEM_DATA[material.Id].Size * material.Amount * batchAmount
	}

	// Add to net size for each output
	netSize += ITEM_DATA[recipeData.Output.Id].Size * recipeData.Output.Amount * batchAmount

	return netSize
}

func (recipeData *RecipeData) CreateOutput() Item {
	output := recipeData.Output
	itemData := ITEM_DATA[output.Id]
	if output.Durability == RECIPE_OUTPUT_MAX_DURABILITY {
		output.Durability = itemData.GetMaxDurability()
	}

	return output
}
