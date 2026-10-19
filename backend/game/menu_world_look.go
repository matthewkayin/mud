package game

import (
	"fmt"
	"strings"
	"strconv"
	"mud/world"
	"mud/util"
)

var MENU_WORLD_ENTRIES_LOOK = map[string]MenuEntry {
	"look": {
		usage: "look",
		description: "Describe the current room",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			room := &gamestate.world.Rooms[playerMob.Data.Room]
			*player.inbox <- fmt.Sprintf("You are in %s.", room.Name)
			describeRoomToPlayer(gamestate, player, room)
			return true
		},
	},

	"search": {
		usage: "search [container]",
		description: "Search the room or a container for items.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			room := &gamestate.world.Rooms[playerMob.Data.Room]

			targetInventory := &room.Inventory

			// Check for container
			if len(args) != 0 {
				chestInventory, _, err := fuzzyFindChestInventory(room, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				targetInventory = chestInventory
			}

			itemNames := make([]string, 0, len(targetInventory.Items))
			for index := range targetInventory.Length() {
				itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", targetInventory.Items[index].GetNameWithAmount()))
			}

			var itemsString string
			if len(itemNames) == 0 {
				itemsString = "nothing"
			} else {
				itemsString = util.CombineNames(itemNames)
			}

			*player.inbox <- fmt.Sprintf("You see %s.", itemsString)
			return true
		},
	},

	"inspect": {
		usage: "inspect <target> [in <container>]",
		description: "Get information about a player, monster, or item",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Get mob and room handle
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Check for in keyword
			targetWords, containerWords, userSpecifiedIn := splitArgsBy(args, "in")

			// Check container
			targetInventory := &playerMob.Data.Inventory
			chestName := "your inventory"
			if userSpecifiedIn {
				var err error
				targetInventory, chestName, err = fuzzyFindChestInventory(playerRoom, containerWords)
				if err != nil {
					*player.inbox <- err.Error()
				}
			}

			// Get item
			itemIndex := fuzzyFindInventoryItemIndex(targetInventory, targetWords)

			// Handle errors, but only if user specifically specified "in" (i.e. they are inspecting an item and not a target)
			if itemIndex < 0 && userSpecifiedIn {
				switch itemIndex {
					case FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED:
						*player.inbox <- "You must specify an item to inspect."
					case FUZZY_FIND_RESULT_AMBIGUOUS:
						*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in %s.",
							strings.Join(targetWords, " "), chestName)
					case FUZZY_FIND_RESULT_NOT_FOUND:
						*player.inbox <- fmt.Sprintf("There is no item named '%s' in %s.",
							strings.Join(targetWords, " "), chestName)
					case FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE:
						*player.inbox <- fmt.Sprintf("There is no item matching that number in %s.", chestName)
				}

				return true
			}

			// Inspect item
			if itemIndex >= 0 {
				item := &targetInventory.Items[itemIndex]
				itemData := world.ITEM_DATA[item.Id]

				*player.inbox <- fmt.Sprintf("Item: <g>%s</g>", itemData.Name)
				*player.inbox <- fmt.Sprintf("Description: %s", itemData.Description)
				*player.inbox <- fmt.Sprintf("Type: %s", itemData.Kind.String())
				*player.inbox <- fmt.Sprintf("Size: %d", itemData.Size)

				// Stat requirements
				stats := item.GetStatRequirements()
				if stats != nil {
					parts := make([]string, 0, world.STAT_COUNT)
					for index := range world.STAT_COUNT {
						if stats.Values[index] != 0 {
							parts = append(parts, fmt.Sprintf("%d %s", stats.Values[index], world.STAT_DATA[index].Abbreviation))
						}
					}

					if len(parts) != 0 {
						*player.inbox <- fmt.Sprintf("Requires: %s", util.CombineNames(parts))
					}
				}

				// Stat bonuses
				stats = item.GetStatBonuses()
				if stats != nil {
					parts := make([]string, 0, world.STAT_COUNT)
					for index := range world.STAT_COUNT {
						if stats.Values[index] != 0 {
							parts = append(parts, fmt.Sprintf("%s %s", statBonusStr(stats.Values[index]), world.STAT_DATA[index].Abbreviation))
 						}
					}

					if len(parts) != 0 {
						*player.inbox <- fmt.Sprintf("Requires: %s", util.CombineNames(parts))
					}
				}

				// Item type specific description
				switch itemData.Kind {
					case world.ITEM_KIND_EQUIPMENT_OUTFIT: {
						outfitData := itemData.Data.(*world.ItemDataArmor)
						*player.inbox <- fmt.Sprintf("Armor: %d", outfitData.Armor)
						*player.inbox <- fmt.Sprintf("Stealth Penalty: %d%%", int(outfitData.StealthPenalty * 100))
					}
					case world.ITEM_KIND_EQUIPMENT_ONE_HANDED, world.ITEM_KIND_EQUIPMENT_TWO_HANDED: {
						weaponData := itemData.Data.(*world.ItemDataWeapon)
						*player.inbox <- fmt.Sprintf("Damage: %d", weaponData.Damage)
					}
				}

				return true
			}

			// Get target handle
			targetHandle, err := fuzzyFindTarget(gamestate, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// TODO: Handle NPC case for friendly NPC interactions

			// Handle monster case
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.Npc != nil {
				// General description
				*player.inbox <- fmt.Sprintf("<y>%s</y> (Level %d): %s",
					targetMob.GetName(), targetMob.Data.Level, targetMob.Npc.GetDescription())

				// Status description
				statusDescription, hasStatusDescription := targetMob.Npc.GetStatusDescription(gamestate.world)
				if hasStatusDescription {
					*player.inbox <- statusDescription
				}

				return true
			}

			// Handle player case
			*player.inbox <- fmt.Sprintf("<y>%s</y> (Level %d): %s %s %s.",
				targetMob.Data.Name, targetMob.Data.Level,
				world.RACE_DATA[targetMob.PlayerCharacter.Race].Name,
				world.CLASS_DATA[targetMob.PlayerCharacter.Class].Name,
				world.JOB_DATA[targetMob.PlayerCharacter.Job].Name)
			printMobEquipmentList(player, targetMob)

			return true
		},
	},

	"hp": {
		usage: "hp [<player>] friends",
		description: "Show your combat status including HP, MP, and conditions. Type 'hp <player>' to see this status for another player in the room. Type 'hp friends' to see this status for all players in the room.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// Determine list of targets for the HP table

			targets := make([]world.MobHandle, 0, 1)

			if len(args) == 0 {
				targets = append(targets, player.mobHandle)
			} else if len(args) == 1 && strings.EqualFold(args[0], "friends") {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
				for _, occupantHandle := range playerRoom.Occupants {
					occupantMob := gamestate.world.Mobs.Get(occupantHandle)

					// Skip NPCs
					if !occupantMob.IsPlayer() {
						continue
					}

					targets = append(targets, occupantHandle)
				}
			} else {
				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.IsNpc() {
					*player.inbox <- fmt.Sprintf("Cannot get status of %s because they are not a player.", targetMob.Data.Name)
					return true
				}

				targets = append(targets, targetHandle)
			}

			headers := []string {
				"Name",
				"HP",
				"MP",
				"Conditions",
			}

			rows := make([]string, 0, len(headers) * len(targets))
			for _, target := range targets {
				targetMob := gamestate.world.Mobs.Get(target)
				rows = append(rows, targetMob.Data.Name)
				rows = append(rows, fmt.Sprintf("%d / %d", targetMob.Data.Health, targetMob.Data.MaxHealth()))
				rows = append(rows, fmt.Sprintf("%d / %d", targetMob.Data.Mana, targetMob.Data.MaxMana()))

				if len(targetMob.Data.Conditions) == 0 {
					rows = append(rows, "None")
				} else {
					conditionNames := make([]string, 0, len(targetMob.Data.Conditions))
					for index := range len(targetMob.Data.Conditions) {
						conditionData := world.CONDITION_DATA[targetMob.Data.Conditions[index].Id]
						conditionNames = append(conditionNames, conditionData.Name)
					}

					rows = append(rows, util.CombineNames(conditionNames))
				}
			}

			printTable(player, "Status", headers, rows)

			return true
		},
	},

	"stats": {
		usage: "stats",
		description: "Show your current stats.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

			sections := [][]string {
				{
					"Race", world.RACE_DATA[player.character.Race].Name,
					"Class", world.CLASS_DATA[player.character.Class].Name,
					"Job", world.JOB_DATA[player.character.Job].Name,
				},
				{
					"Level", strconv.Itoa(int(playerMob.Data.Level)),
					"Experience", fmt.Sprintf("%d / %d", playerMob.Data.Experience, playerMob.Data.ExperienceToNextLevel),
				},
				{
					"HP", fmt.Sprintf("%d / %d", playerMob.Data.Health, playerMob.Data.MaxHealth()),
					"MP", fmt.Sprintf("%d / %d", playerMob.Data.Mana, playerMob.Data.MaxMana()),
				},
				{
					"Armor", fmt.Sprintf("%d", playerMob.Data.Armor()),
					"Stealth Penalty", fmt.Sprintf("%.2f", playerMob.GetStealthPenalty()),
				},
			}

			statBonuses := &playerMob.Data.Equipment.StatBonuses
			statRows := make([]string, 0, 4 * world.STAT_COUNT)

			for index := range world.STAT_COUNT {
				statRows = append(statRows, world.STAT_DATA[index].Name)
				statRows = append(statRows, strconv.Itoa(int(playerMob.Data.Stats.Values[index])))
				statRows = append(statRows, statModifierStr(statBonuses.Values[index]))
				statRows = append(statRows, strconv.Itoa(int(playerMob.Data.GetStat(index))))
			}

			printSheetTable(player, playerMob.Data.Name, sections, []string{"Stat", "Base", "Equipment", "Total"}, statRows)

			abilities := playerMob.Data.GetAbilityList()
			if len(abilities) != 0 {
				abilityRows := make([]string, 0, 2 * len(abilities))
				for _, ability := range abilities {
					abilityData := world.MOB_ABILITY_DATA[ability]
					abilityRows = append(abilityRows, abilityData.Name)
					abilityRows = append(abilityRows, abilityData.Description)
				}

				*player.inbox <- ""
				printTable(player, "Talents", []string{"Name", "Description"}, abilityRows)
			} else {
				*player.inbox <- "\nYou have no talents."
			}

			return true
		},
	},
}
