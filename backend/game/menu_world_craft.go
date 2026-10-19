package game

import (
	"fmt"
	"strings"
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

	"cast": {
		usage: "cast <spell> [on <target>]",
		description: "Cast a spell",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			spellWords, targetWords, userSpecifiedTarget := splitArgsBy(args, "on")

			if len(spellWords) == 0 || (userSpecifiedTarget && len(targetWords) == 0) {
				return false
			}

			// Find the spell in their spell list
			spell, err := fuzzyFindCastableSpell(gamestate, player, spellWords)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Find the target in the room
			target, err := findSpellTarget(gamestate, player, spell, userSpecifiedTarget, targetWords)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check for protected NPC
			if target.Type == world.MOB_TARGET_MOB {
				targetHandle := target.Value.(world.MobHandle)
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.Npc.IsProtected() {
					*player.inbox <- fmt.Sprintf("You cannot attack %s, they are a friend!", targetMob.Data.Name)
					return true
				}
			}

			// Queue up a cast action
			player.nextAction = Action {
				actionType: ACTION_TYPE_CAST,
				data: ActionCast {
					spell: spell,
					target: target,
				},
			}

			return true
		},
	},
}
