package game

import (
	"fmt"
	"strings"
	"strconv"
	"mud/world"
	"mud/util"
	"mud/bitset"
)

var MENU_WORLD_ENTRIES_LOOK = map[string]MenuEntry {
	"look": {
		usage: "look",
		description: "Describe the current room",
		groups: []string { "exploration" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			room := &gamestate.world.Rooms[playerMob.Data.Room]
			player.send(fmt.Sprintf("You are in %s.", room.Name))
			describeRoomToPlayer(gamestate, player, room)
			return true
		},
	},

	"exits": {
		usage: "exits",
		description: "Describe the exits of the current room",
		groups: []string { "exploration" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			room := &gamestate.world.Rooms[playerMob.Data.Room]

			// Exits
			exitFound := false
			for directionIndex := range world.DIRECTION_COUNT {
				direction := world.Direction(directionIndex)

				if room.Exits[direction] == world.ROOM_NONE {
					continue
				}

				roomName := "an undiscovered room"
				if bitset.Check(player.character.RoomsDiscovered, room.Exits[direction]) {
					roomName = gamestate.world.Rooms[room.Exits[direction]].Name
				}

				player.send(fmt.Sprintf("To the <p>%s</p> is %s.", direction.String(), roomName))
				exitFound = true
			}
			if !exitFound {
				player.send("This room has no exits!")
			}

			return true
		},
	},

	"search": {
		usage: "search [container]",
		description: "Search the room or a container for items.",
		groups: []string { "exploration", "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			room := &gamestate.world.Rooms[playerMob.Data.Room]

			targetInventory := &room.Inventory

			// Check for container
			if len(args) != 0 {
				chestInventory, _, err := fuzzyFindChestInventory(room, args)
				if err != nil {
					player.send(err.Error())
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

			player.send(fmt.Sprintf("You see %s.", itemsString))
			return true
		},
	},

	"inspect": {
		usage: "inspect <target> [in <container>]",
		description: "Get information about a player, monster, or item",
		groups: []string { "exploration" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Get mob and room handle
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Check for 'in' keyword
			targetWords, containerWords, userSpecifiedIn := splitArgsBy(args, "in")

			// Check container
			targetInventory := &playerMob.Data.Inventory
			chestName := "your inventory"
			if userSpecifiedIn {
				var err error
				targetInventory, chestName, err = fuzzyFindChestInventory(playerRoom, containerWords)
				if err != nil {
					player.send(err.Error())
					return true
				}
			}

			// Get item
			itemIndex := fuzzyFindInventoryItemIndex(targetInventory, targetWords)

			// Handle errors, but only if user specifically specified "in" (i.e. they are inspecting an item and not a target)
			if itemIndex < 0 && userSpecifiedIn {
				switch itemIndex {
					case FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED:
						player.send("You must specify an item to inspect.")
					case FUZZY_FIND_RESULT_AMBIGUOUS:
						player.send(fmt.Sprintf("There are multiple items matching '%s' in %s.",
							strings.Join(targetWords, " "), chestName))
					case FUZZY_FIND_RESULT_NOT_FOUND:
						player.send(fmt.Sprintf("There is no item named '%s' in %s.",
							strings.Join(targetWords, " "), chestName))
					case FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE:
						player.send(fmt.Sprintf("There is no item matching that number in %s.", chestName))
				}

				return true
			}

			// Inspect item
			if itemIndex >= 0 {
				item := &targetInventory.Items[itemIndex]
				itemData := world.ITEM_DATA[item.Id]

				player.send(fmt.Sprintf("Item: <g>%s</g>", item.GetNameWithCondition()))
				player.send(fmt.Sprintf("Description: %s", itemData.Description))
				player.send(fmt.Sprintf("Type: %s", itemData.Kind.String()))
				player.send(fmt.Sprintf("Size: %d", itemData.Size))

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
						player.send(fmt.Sprintf("Requires: %s", util.CombineNames(parts)))
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
						player.send(fmt.Sprintf("Requires: %s", util.CombineNames(parts)))
					}
				}

				// Item type specific description
				switch itemData.Kind {
					case world.ITEM_KIND_EQUIPMENT_OUTFIT, world.ITEM_KIND_EQUIPMENT_SHIELD: {
						outfitData := itemData.Data.(*world.ItemDataArmor)
						player.send(fmt.Sprintf("Durability: %d", item.GetDurabilityPercent()))
						player.send(fmt.Sprintf("Deflection: %d", outfitData.Deflection))
						player.send(fmt.Sprintf("Stealth Penalty: %d%%", int(outfitData.StealthPenalty * 100)))
					}
					case world.ITEM_KIND_EQUIPMENT_ONE_HANDED, world.ITEM_KIND_EQUIPMENT_TWO_HANDED: {
						player.send(fmt.Sprintf("Durability: %d%%", item.GetDurabilityPercent()))
						player.send(fmt.Sprintf("Damage: %d", item.GetDamage()))
						player.send(fmt.Sprintf("Penetration: %d", item.GetArmorPenetration()))
					}
				}

				return true
			}

			// Get target handle
			targetHandle, err := fuzzyFindTarget(gamestate, player, args)
			if err == nil {
				// Handle monster case
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.Npc != nil {
					// General description
					player.send(fmt.Sprintf("<y>%s</y> (Level %d): %s",
						targetMob.GetName(), targetMob.Data.Level, targetMob.Npc.GetDescription()))


					// Status description
					statusDescription, hasStatusDescription := targetMob.Npc.GetStatusDescription(gamestate.world)
					if hasStatusDescription {
						player.send(statusDescription)
					}

					// Are they friendly?!
					var dispositionMsg string
					switch targetMob.Npc.GetDisposition() {
						case world.NPC_DISPOSITION_FRIENDLY:
							dispositionMsg = fmt.Sprintf("%s seems to be friendly", targetMob.GetName())
							if targetMob.Npc.IsMerchant() {
								dispositionMsg += ", and appears to be in business!"
							} else {
								dispositionMsg += "."
							}
						case world.NPC_DISPOSITION_HOSTILE:
							dispositionMsg = fmt.Sprintf("%s appears to be very unfriendly.", targetMob.GetName())
						default:
							dispositionMsg = fmt.Sprintf("It is not clear if %s is friendly or unfriendly.", targetMob.GetName())
					}
					player.send(dispositionMsg)

					return true
				}

				// Handle player case
				player.send(fmt.Sprintf("<y>%s</y> (Level %d): %s %s %s.",
					targetMob.Data.Name, targetMob.Data.Level,
					world.RACE_DATA[targetMob.PlayerCharacter.Race].Name,
					world.CLASS_DATA[targetMob.PlayerCharacter.Class].Name,
					world.JOB_DATA[targetMob.PlayerCharacter.Job].Name))
				printMobEquipmentList(player, targetMob)
			}

			// Handle fixture case
			fixture, err := fuzzyFindFixture(gamestate, player, args)
			if err == nil {
				description, ok := fixture.GetDescription(gamestate.world)
				if !ok {
					player.send(fmt.Sprintf("%s is undescribable.", fixture.GetName()))
					return true
				}

				// Send description
				player.send(description)
				player.send(fmt.Sprintf("You can interact with %s by typing %s", fixture.GetName(), util.CombineNames(fixture.GetVerbs())))
					return true
			}

			player.send(err.Error())
			return true
		},
	},

	"hp": {
		usage: "hp [<player>] friends",
		description: "Show your combat status including HP, MP, and conditions. Type 'hp <player>' to see this status for another player in the room. Type 'hp friends' to see this status for all players in the room.",
		groups: []string { "combat" },
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
					player.send(err.Error())
					return true
				}

				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.IsNpc() {
					player.send(fmt.Sprintf("Cannot get status of %s because they are not a player.", targetMob.Data.Name))
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
		groups: []string { "information" },
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
					"Deflection", fmt.Sprintf("%d", playerMob.Data.Deflection()),
					"Damage Resistance", fmt.Sprintf("%d", playerMob.Data.Deflection()),
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

				player.send("")
				printTable(player, "Talents", []string{"Name", "Description"}, abilityRows)
			} else {
				player.send("\nYou have no talents.")
			}

			return true
		},
	},
}
