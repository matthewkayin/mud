package world

import (
	"fmt"

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

// A recipe script whose name has been read, but which hasn't been parsed yet
type recipeScript struct {
	path string
	table *lua.Table
}

var RECIPE_DATA []*RecipeData
var RECIPE_NAME_TO_ID map[string]RecipeId

// Relative to the world data folder
const RECIPE_DATA_FOLDER = "recipes"

// LOAD

// Loads each recipe script and maps its name to a RecipeId, without parsing the rest of the recipe.
// This lets items reference recipes by name before the recipes themselves are parsed.
// The returned scripts are indexed by RecipeId and should be passed to loadRecipeData once items are loaded.
func (world *World) loadRecipeScripts() ([]recipeScript, error) {
	// Read recipe folder
	paths, err := scriptGetFilesFrom(world.dataFolder, RECIPE_DATA_FOLDER)
	if err != nil {
		return nil, fmt.Errorf("Error opening recipe data folder: %s", err.Error())
	}

	recipeScripts := make([]recipeScript, 0, len(paths))
	RECIPE_DATA = make([]*RecipeData, len(paths))
	RECIPE_NAME_TO_ID = make(map[string]RecipeId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return nil, err
		}

		// Get recipe name
		parser := ScriptParser{}
		recipeName := parser.getString(table, "name")
		if len(parser.problems) != 0 {
			return nil, fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateRecipeName := RECIPE_NAME_TO_ID[recipeName]
		if duplicateRecipeName {
			return nil, fmt.Errorf("Recipe %s has name '%s' which is a duplicate of another recipe.", path, recipeName)
		}

		// Store recipe script
		RECIPE_NAME_TO_ID[recipeName] = RecipeId(len(recipeScripts))
		recipeScripts = append(recipeScripts, recipeScript {
			path: path,
			table: table,
		})
	}

	return recipeScripts, nil
}

// Parses the recipe scripts returned by loadRecipeScripts. Must be called after loadItemData.
func (world *World) loadRecipeData(recipeScripts []recipeScript) error {
	for index, script := range recipeScripts {
		// Parse recipe data
		parser := ScriptParser{}
		recipeData := parser.parseRecipe(script.table)
		if recipeData == nil {
			return fmt.Errorf("%s: %s", script.path, parser.getError().Error())
		}

		// Store recipe in RECIPE_DATA
		RECIPE_DATA[index] = recipeData
	}

	return nil
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

			itemId, amount := parser.parseItem(materialTable, key)
			recipeData.Materials = append(recipeData.Materials, RecipeMaterial {
				Id: itemId,
				Amount: amount,
			})
		}
	}

	// Output
	outputTable := parser.getTable(table, "output")
	if outputTable != nil {
		itemId, amount := parser.parseItem(outputTable, "output")
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
