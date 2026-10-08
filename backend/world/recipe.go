package world

import (
	"fmt"
	"log"

	"github.com/mmcdole/lunar"
)

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
var RECIPE_NAME_TO_ID map[string]RecipeId

const RECIPE_DATA_FOLDER = WORLD_DATA_FOLDER + "/recipes"

// LOAD

// Loads each recipe script and maps its name to a RecipeId, without parsing the rest of the recipe.
// This lets items reference recipes by name before the recipes themselves are parsed.
// The returned tables are indexed by RecipeId and should be passed to loadRecipeData once items are loaded.
func (world *World) loadRecipeTables() []*lua.Table {
	// Read recipe folder
	log.Printf("Loading recipe tables...")
	paths, err := scriptGetFilesFrom(RECIPE_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening recipe data folder: %s", err.Error())
	}

	recipeTables := make([]*lua.Table, 0, len(paths))
	RECIPE_DATA = make([]*RecipeData, len(paths))
	RECIPE_NAME_TO_ID = make(map[string]RecipeId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Get recipe name
		parser := ScriptParser{}
		recipeName := parser.getString(table, "name")
		if len(parser.problems) != 0 {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateRecipeName := RECIPE_NAME_TO_ID[recipeName]
		if duplicateRecipeName {
			log.Fatalf("Recipe %s has name '%s' which is a duplicate of another recipe.", path, recipeName)
		}

		// Store recipe table
		RECIPE_NAME_TO_ID[recipeName] = RecipeId(len(recipeTables))
		recipeTables = append(recipeTables, table)
	}

	log.Printf("All recipe tables have been loaded.")
	return recipeTables
}

// Parses the recipe tables returned by loadRecipeTables. Must be called after loadItemData.
func (world *World) loadRecipeData(recipeTables []*lua.Table) {
	log.Printf("Loading recipe data...")

	for index, table := range recipeTables {
		// Parse recipe data
		parser := ScriptParser{}
		recipeData := parser.parseRecipe(table)
		if recipeData == nil {
			recipeName, _ := table.RawGetString("name").AsString()
			log.Fatalf("Recipe '%s': %s", recipeName, parser.getError().Error())
		}

		// Store recipe in RECIPE_DATA
		RECIPE_DATA[index] = recipeData
		log.Printf("Loaded recipe '%s'.", recipeData.Name)
	}

	log.Printf("All recipe data has been loaded.")
}

func (parser *ScriptParser) parseRecipe(table *lua.Table) *RecipeData {
	recipeData := &RecipeData{}

	recipeData.Name = parser.getString(table, "name")

	jobName := parser.getString(table, "job")
	var err error
	recipeData.Job, err = JobIdFromString(jobName)
	parser.addProblem(err)

	recipeData.Level = parser.getInt32(table, "level")
	if recipeData.Level <= 0 {
		parser.addProblem(fmt.Errorf("field 'level' must be greater than 0, got %d", recipeData.Level))
	}

	// Materials
	materialsTable := parser.getTable(table, "materials")
	if materialsTable != nil {
		materialCount := materialsTable.RawLen()
		if materialCount == 0 {
			parser.addProblem(fmt.Errorf("field 'materials' must not be empty"))
		}

		recipeData.Materials = make([]RecipeMaterial, 0, materialCount)
		for index := 1; index <= materialCount; index++ {
			key := fmt.Sprintf("materials[%d]", index)
			materialTable, ok := materialsTable.RawGetInt(index).AsTable()
			if !ok {
				parser.addProblem(fmt.Errorf("field '%s' must be a table", key))
				continue
			}

			itemId, amount := parser.parseRecipeItem(materialTable, key)
			recipeData.Materials = append(recipeData.Materials, RecipeMaterial {
				Id: itemId,
				Amount: amount,
			})
		}
	}

	// Output
	outputTable := parser.getTable(table, "output")
	if outputTable != nil {
		itemId, amount := parser.parseRecipeItem(outputTable, "output")
		recipeData.Output = Item {
			Id: itemId,
			Amount: amount,
			Durability: RECIPE_OUTPUT_MAX_DURABILITY,
		}
	}

	if len(parser.problems) != 0 {
		return nil
	}

	return recipeData
}

// Parses a table of the form { item = "<item name>", amount = N }
func (parser *ScriptParser) parseRecipeItem(table *lua.Table, key string) (ItemId, int32) {
	itemName := parser.getString(table, "item")
	itemId, exists := ITEM_NAME_TO_ID[itemName]
	if !exists {
		parser.addProblem(fmt.Errorf("field '%s' references item '%s' which does not exist", key, itemName))
	}

	amount := parser.getInt32(table, "amount")
	if amount <= 0 {
		parser.addProblem(fmt.Errorf("field '%s' amount must be greater than 0, got %d", key, amount))
	}

	return itemId, amount
}

// HELPERS

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
