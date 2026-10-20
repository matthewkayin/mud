package game

import (
	"fmt"
	"strings"
	"strconv"
	"mud/world"
)

var MENU_WORLD_ENTRIES_CRAFT = map[string]MenuEntry {
	"recipe": {
		usage: "recipe [list] [info <recipe>]",
		description: "Get information about your known recipes.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

			if args[0] == "list" {
				if len(player.character.RecipesKnown) == 0 {
					*player.inbox <- "You do not know any crafting recipes."
					return true
				}
				rows := make([]string, 0, 3 * len(player.character.RecipesKnown))
				for _, recipe := range player.character.RecipesKnown {
					recipeData := world.RECIPE_DATA[recipe]

					materialNames := make([]string, 0, len(recipeData.Materials))
					for _, material := range recipeData.Materials {
						materialNames = append(materialNames, fmt.Sprintf("%s x%d", world.ITEM_DATA[material.Id].Name, material.Amount))
					}

					rows = append(rows, recipeData.Name)
					rows = append(rows, fmt.Sprintf("%d", recipeData.Level))
					rows = append(rows, strings.Join(materialNames, ", "))
				}

				printTable(player, "Known Recipes", []string{"Name", "Level", "Materials"}, rows)
				return true
			}

			if args[0] == "info" {
				query, err := fuzzyFindKnownRecipe(player.character, args[1:])

				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				recipeData := world.RECIPE_DATA[query]
				rows := make([]string, 0, 3 * len(recipeData.Materials))
				for _, ingredient := range recipeData.Materials {
					rows = append(rows, world.ITEM_DATA[ingredient.Id].Name)
					rows = append(rows, fmt.Sprintf("%d", playerMob.Data.Inventory.AmountOf(ingredient.Id)))
					rows = append(rows, fmt.Sprintf("%d", ingredient.Amount))
				}

				printTable(player, recipeData.Name, []string{"Material", "Have", "Needed"}, rows)

				return true
			}

			return false
		},
	},

	"craft": {
		usage: "craft [<amount>] <item>",
		description: "Craft an item for which you know the recipe",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			// Check if they are in comabt
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.IsInCombat(gamestate.world) {
				*player.inbox <- "You cannot craft items while you are in combat."
				return true
			}

			//check whether they put in a quantity
			amount, atoiErr := strconv.Atoi(args[0])
			var batchAmount int32 = 1
			if atoiErr == nil {
				args = args[1:]
				batchAmount = int32(amount)
			}

			recipe, fuzErr := fuzzyFindKnownRecipe(player.character, args)
			if fuzErr != nil {
				*player.inbox <- fuzErr.Error()
				return true
			}

			err := playerMob.CanCraft(recipe, batchAmount)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			player.nextAction = Action {
				actionType: ACTION_TYPE_CRAFT_ITEM,
				data: ActionCraftItem {
					target: recipe,
					amount: batchAmount,
				},
			}

			return true
		},
	},
}
