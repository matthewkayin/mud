package game

import (
	"fmt"
	"log"
	"slices"
	"mud/bitset"
	"mud/world"
	"strings"
)

var MENU_WORLD = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		// Clear the player's action in case they had any leftover from a previous login session
		player.nextAction = Action {
			actionType: ACTION_TYPE_NONE,
			data: nil,
		}

		// Create a mob for the player
		playerMob := world.MobInitFromCharacter(player.character)
		player.mobHandle = gamestate.world.Mobs.Push(playerMob)
		playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
		gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s has joined the room.", player.character.Data.Name))
		playerRoom.AddOccupant(gamestate.world, player.mobHandle)

		// Enter world menu
		*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", player.character.Data.Name)
		room := &gamestate.world.Rooms[player.character.Data.Room]
		describeRoomToPlayer(gamestate, player, room)
	},
	onExit: func(gamestate *GameState, player *Player) {
		tradeSessionOnPlayerLogout(gamestate, player)

		// Get player mob
		// Note that player mob will not exist here if the player died
		playerMob, mobExists := gamestate.world.Mobs.GetIfExists(player.mobHandle)
		if mobExists {
			// Remove player from current room
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
			playerRoom.RemoveOccupant(player.mobHandle)
			gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s has left the world.", playerMob.Data.Name))

			// Save player mob data back to their character
			player.character.Data = playerMob.Data
			world.SaveCharacter(player.character)
		}

		player.character = nil
	},

	entries: map[string]MenuEntry {
		"logout": {
			usage: "logout",
			description: "Logout of the world",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)
				return true
			},
		},

		"who": {
			usage: "who",
			description: "Get a list of everyone who is online",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				names := make([]string, 0, len(gamestate.players))
				for index := range len(gamestate.players) {
					if !gamestate.players[index].isLoggedIn() {
						continue
					}

					names = append(names, gamestate.players[index].character.Data.Name)
				}

				*player.inbox <- fmt.Sprintf("The players in the world are: %s.", combineNames(names))
				return true
			},
		},

		"look": {
			usage: "look",
			description: "Describe the current room",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				room := &gamestate.world.Rooms[playerMob.Data.Room]
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
					itemNames = append(itemNames, targetInventory.Items[index].GetNameWithAmount())
				}

				var itemsString string
				if len(itemNames) == 0 {
					itemsString = "nothing"
				} else {
					itemsString = combineNames(itemNames)
				}

				*player.inbox <- fmt.Sprintf("You see %s.", itemsString)
				return true
			},
		},

		"inspect": {
			usage: "inspect <mob>",
			description: "Get information about a player or monster in the room",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Get target handle
				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// TODO: Handle NPC case for friendly NPC interactions

				// Handle monster case
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter == nil {
					*player.inbox <- fmt.Sprintf("%s is a level %d monster. It appears to be hostile!",
						targetMob.GetName(), targetMob.Data.Level)
					return true
				}

				// Handle player case
				*player.inbox <- fmt.Sprintf("%s is a level %d %s %s %s.",
					targetMob.Data.Name, targetMob.Data.Level,
					world.RACE_DATA[targetMob.PlayerCharacter.Race].Name,
					world.CLASS_DATA[targetMob.PlayerCharacter.Class].Name,
					world.JOB_DATA[targetMob.PlayerCharacter.Job].Name)
				*player.inbox <- "They are wearing the following:"
				printMobEquipmentList(player, targetMob)

				return true
			},
		},

		"say": {
			usage: "say <message>",
			description: "Send a messsage to the current room",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 1 {
					return false
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s said '%s'",
					player.character.Data.Name,
					strings.TrimSpace(strings.Join(args, " "))))
				return true
			},
		},

		"tell": {
			usage: "tell <target> <message>",
			description: "Send a message to a specific person in the current room",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Parse input
				argString := strings.Join(args, " ")
				targetString, messageString, colonFound := strings.Cut(argString, ":")
				message := strings.TrimSpace(messageString)
				if !colonFound || len(targetString) == 0 || len(message) == 0 {
					return false
				}

				// Find target
				targetHandle, err := fuzzyFindTarget(gamestate, player, strings.Fields(targetString))
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Handle case where they talk to themselves
				if targetHandle == player.mobHandle {
					*player.inbox <- fmt.Sprintf("You told yourself: '%s'", message)
					return true
				}

				// Handle case where they talk to a player
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter != nil {
					targetPlayerIndex, targetPlayerExists := gamestate.playerIdToIndexMap[targetMob.PlayerCharacter.PlayerId]
					if !targetPlayerExists {
						log.Printf("Warn - Tried to tell player %d a message, but they don't exist.", targetMob.PlayerCharacter.PlayerId)
						return true
					}
					targetPlayer := &gamestate.players[targetPlayerIndex]

					*player.inbox <- fmt.Sprintf("You told %s: '%s'", targetMob.GetName(), message)
					*targetPlayer.inbox <- fmt.Sprintf("%s told you: '%s'", playerMob.GetName(), message)
					return true
				}

				// TODO: handle case where they talk to an NPC

				return true
			},
		},

		"yell": {
			usage: "yell <message>",
			description: "Send a message to everyone in this and the adjacent rooms",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				// Get a handle to the player room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Collect a list containing all rooms to broadcast to
				rooms := make([]int, 0, 5)
				rooms = append(rooms, playerMob.Data.Room)

				// Add room exits to the list
				for direction := range world.DIRECTION_COUNT {
					roomIndex := playerRoom.Exits[direction]
					if roomIndex == world.ROOM_NONE {
						continue
					}

					rooms = append(rooms, roomIndex)
				}

				// Broadcast the message to each room
				message := fmt.Sprintf("%s: '%s'",
					playerMob.Data.Name,
					strings.TrimSpace(strings.Join(args, " ")))
				for _, roomIndex := range rooms {
					gamestate.messageRoom(roomIndex, message)
				}

				return true
			},
		},

		"move": {
			usage: "move <direction>",
			description: "Walk to an adjacent room.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 1 {
					return false
				}

				// Get the player mob and room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Determine the index of the target room
				direction, directionFound := world.DirectionFromString(args[0])
				if !directionFound {
					*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
					return true
				}

				// Move player
				err := playerRoom.MoveOccupant(gamestate.world, player.mobHandle, direction)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				playerRoom = &gamestate.world.Rooms[playerMob.Data.Room]
				*player.inbox <- fmt.Sprintf("You moved into %s.", playerRoom.Name)
				describeRoomToPlayer(gamestate, player, playerRoom)
				return true
			},
		},

		"hp": {
			usage: "hp [<player>] friends",
			description: "Show your combat status including HP, MP, and conditions. Type 'hp <player>' to see this status for another player in the room. Type 'hp friends' to see this status for all players in the room.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					playerMob := gamestate.world.Mobs.Get(player.mobHandle)
					printMobHp(player, playerMob)

					return true
				}

				if len(args) == 1 && strings.EqualFold(args[0], "friends") {
					playerMob := gamestate.world.Mobs.Get(player.mobHandle)
					playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
					for _, occupantHandle := range playerRoom.Occupants {
						occupantMob := gamestate.world.Mobs.Get(occupantHandle)

						// Skip NPCs
						if occupantMob.PlayerCharacter == nil {
							continue
						}

						printMobHp(player, occupantMob)
						*player.inbox <- "\n"
					}

					return true
				}

				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				targetMob := gamestate.world.Mobs.Get(targetHandle)
				printMobHp(player, targetMob)

				return true
			},
		},

		"stats": {
			usage: "stats",
			description: "Show your current stats.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				classData := world.CLASS_DATA[player.character.Class]
				raceData := world.RACE_DATA[player.character.Race]

				*player.inbox <- fmt.Sprintf("%s - Level %d %s %s",
					playerMob.Data.Name, playerMob.Data.Level, raceData.Name, classData.Name)
				*player.inbox <- fmt.Sprintf("Experience: %d / %d", playerMob.Data.Experience, playerMob.Data.ExperienceToNextLevel)

				*player.inbox <- fmt.Sprintf("\nHP: %d / %d", playerMob.Data.Health, playerMob.Data.MaxHealth())
				*player.inbox <- fmt.Sprintf("MP: %d / %d", playerMob.Data.Mana, playerMob.Data.MaxMana())

				statBonuses := &playerMob.Data.Equipment.StatBonuses
				*player.inbox <- fmt.Sprintf("\nVitality: %d (+%d)", playerMob.Data.Stats.Vitality, statBonuses.Vitality)
				*player.inbox <- fmt.Sprintf("Strength: %d (+%d)", playerMob.Data.Stats.Strength, statBonuses.Strength)
				*player.inbox <- fmt.Sprintf("Agility: %d (+%d)", playerMob.Data.Stats.Agility, statBonuses.Agility)
				*player.inbox <- fmt.Sprintf("Intelligence: %d (+%d)", playerMob.Data.Stats.Intelligence, statBonuses.Intelligence)
				*player.inbox <- fmt.Sprintf("Faith: %d (+%d)", playerMob.Data.Stats.Faith, statBonuses.Faith)

				return true
			},
		},

		"stop": {
			usage: "stop",
			description: "Stop attacking or cancel your current spell",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				switch playerMob.Mode {
					case world.MOB_MODE_IDLE: {}
					case world.MOB_MODE_ATTACK: {
						targetMob, targetExists := gamestate.world.Mobs.GetIfExists(playerMob.Target)
						if targetExists {
							gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s stopped attacking %s", playerMob.Data.Name, targetMob.Data.Name))
						}
					}
					case world.MOB_MODE_CAST: {
						gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s canceled their spell.", playerMob.Data.Name))
					}
				}

				playerMob.Mode = world.MOB_MODE_IDLE
				player.nextAction = Action {
					actionType: ACTION_TYPE_NONE,
					data: nil,
				}

				return true
			},
		},

		"attack": {
			usage: "attack <target>",
			description: "Attack the target.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 1 {
					return false
				}

				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Check for PvP
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter != nil {
					*player.inbox <- "You cannot attack other adventurers!"
					return true
				}

				player.nextAction = Action {
					actionType: ACTION_TYPE_ATTACK,
					data: ActionAttack {
						target: targetHandle,
					},
				}

				return true
			},
		},

		"inventory": {
			usage: "inventory",
			description: "List the items in your inventory",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				inventorySize := len(playerMob.Data.Inventory.Items)

				if playerMob.Data.Inventory.Length() == 0 {
					*player.inbox <- "There is nothing in your inventory."
					return true
				}

				for index := range inventorySize {
					item := &playerMob.Data.Inventory.Items[index]
					itemData := world.ITEM_DATA[item.Id]
					itemStatRequirements := item.GetStatRequirements()

					typeStr := world.ItemTypeToString(itemData.ItemType)
					if itemStatRequirements != nil {
						statStrings := make([]string, 0, 5)
						if itemStatRequirements.Vitality != 0 {
							statStrings = append(statStrings, fmt.Sprintf("VIT %d", itemStatRequirements.Vitality))
						}
						if itemStatRequirements.Strength != 0 {
							statStrings = append(statStrings, fmt.Sprintf("STR %d", itemStatRequirements.Strength))
						}
						if itemStatRequirements.Agility != 0 {
							statStrings = append(statStrings, fmt.Sprintf("AGI %d", itemStatRequirements.Agility))
						}
						if itemStatRequirements.Intelligence != 0 {
							statStrings = append(statStrings, fmt.Sprintf("INT %d", itemStatRequirements.Intelligence))
						}
						if itemStatRequirements.Faith != 0 {
							statStrings = append(statStrings, fmt.Sprintf("FTH %d", itemStatRequirements.Faith))
						}
						if len(statStrings) != 0 {
							typeStr += fmt.Sprintf(" (Requires %s)", strings.Join(statStrings, ", "))
						}
					}

					itemName := item.GetNameWithCondition()
					if item.Amount > 1 {
						itemName = fmt.Sprintf("%s (x%d)", itemName, item.Amount)
					}
					*player.inbox <- fmt.Sprintf("%s | Type: %s | Description: %s", itemName, typeStr, itemData.Description)
				}
				itemNames := make([]string, 0, inventorySize)
				for _, item := range playerMob.Data.Inventory.Items {
					itemNames = append(itemNames, world.ITEM_DATA[item.Id].Name)
				}
				return true
			},
		},

		"drop": {
			usage: "drop <item>",
			description: "Drop an item from your inventory",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				result := inventoryTransfer(&playerMob.Data.Inventory, &playerRoom.Inventory, args)
				switch result.status {
					case INVENTORY_TRANSFER_STATUS_PARTIAL:
						*player.inbox <- fmt.Sprintf("You only have %d %s in your inventory.", result.amount, result.itemName)
						fallthrough
					case INVENTORY_TRANSFER_STATUS_OK:
						*player.inbox <- fmt.Sprintf("You dropped %s.", itemNameWithAmount(result.itemName, result.amount))
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED:
						*player.inbox <- "You must specify an item to drop."
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND:
						*player.inbox <- fmt.Sprintf("You have no item named '%s' in your inventory.", result.itemName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS:
						*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in your inventory.", result.itemName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE:
						*player.inbox <- "There is no item matching that number in your inventory."
					case INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK:
						*player.inbox <- fmt.Sprintf("You can only drop 1 %s at once.", result.itemName)
					default:
						panic(fmt.Sprintf("Transfer result status %d not handled.", result.status))
				}

				return true
			},
		},

		"put": {
			usage: "put <item> into <container>",
			description: "Puts an item from your inventory into the specified container",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				room := &gamestate.world.Rooms[playerMob.Data.Room]

				// Split args
				itemWords, chestWords, userSpecifiedInto := splitArgsBy(args, "into")
				if !userSpecifiedInto || len(itemWords) == 0 || len(chestWords) == 0 {
					return false
				}

				// Find target inventory
				targetInventory, chestName, err := fuzzyFindChestInventory(room, chestWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Transfer
				result := inventoryTransfer(&playerMob.Data.Inventory, targetInventory, itemWords)
				switch result.status {
					case INVENTORY_TRANSFER_STATUS_PARTIAL:
						*player.inbox <- fmt.Sprintf("You only have %d %s in your inventory.", result.amount, result.itemName)
						fallthrough
					case INVENTORY_TRANSFER_STATUS_OK:
						*player.inbox <- fmt.Sprintf("You put %s into %s.", itemNameWithAmount(result.itemName, result.amount), chestName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED:
						*player.inbox <- "You must specify an item to put."
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND:
						*player.inbox <- fmt.Sprintf("You have no item named '%s' in your inventory.", result.itemName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS:
						*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in your inventory.", result.itemName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE:
						*player.inbox <- "There is no item matching that number in your inventory."
					case INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK:
						*player.inbox <- fmt.Sprintf("You can only put 1 %s at once.", result.itemName)
					default:
						panic(fmt.Sprintf("Transfer result status %d not handled.", result.status))
				}

				return true
			},
		},

		"give": {
			usage: "give <item> to <target>",
			description: "Gives an item to a player",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				// Split args
				itemWords, targetWords, userSpecifiedTo := splitArgsBy(args, "to")
				if !userSpecifiedTo || len(itemWords) == 0 || len(targetWords) == 0 {
					return false
				}

				// Find target
				targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// TODO: allow giving to non-player NPCs for things like RP and encounters?
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter == nil {
					*player.inbox <- "You cannot give an item to someone who isn't a player."
					return true
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				result := inventoryTransfer(&playerMob.Data.Inventory, &targetMob.Data.Inventory, itemWords)
				switch result.status {
					case INVENTORY_TRANSFER_STATUS_PARTIAL:
						*player.inbox <- fmt.Sprintf("You only have %d %s in your inventory.", result.amount, result.itemName)
						fallthrough
					case INVENTORY_TRANSFER_STATUS_OK:
						*player.inbox <- fmt.Sprintf("You gave %s to %s.", itemNameWithAmount(result.itemName, result.amount), targetMob.GetName())
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED:
						*player.inbox <- "You must specify an item to give."
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND:
						*player.inbox <- fmt.Sprintf("You have no item named '%s' in your inventory.", result.itemName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS:
						*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in your inventory.", result.itemName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE:
						*player.inbox <- "There is no item matching that number in your inventory."
					case INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK:
						*player.inbox <- fmt.Sprintf("You can only give 1 %s at once.", result.itemName)
					default:
						panic(fmt.Sprintf("Transfer result status %d not handled.", result.status))
				}

				return true
			},
		},

		"take": {
			usage: "take <item> [from <container>]",
			description: "Pick up an item in your current room",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				itemWords, chestWords, userSpecifiedChest := splitArgsBy(args, "from")
				if userSpecifiedChest && len(chestWords) == 0 {
					return false
				}

				// Get pointer to room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Determine target inventory
				targetInventory := &playerRoom.Inventory
				chestName := "the room"
				if userSpecifiedChest {
					var err error
					targetInventory, chestName, err = fuzzyFindChestInventory(playerRoom, chestWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				}

				// Inventory transfer
				result := inventoryTransfer(targetInventory, &playerMob.Data.Inventory, itemWords)
				switch result.status {
					case INVENTORY_TRANSFER_STATUS_PARTIAL:
						*player.inbox <- fmt.Sprintf("There is only %d %s in %s", result.amount, result.itemName, chestName)
						fallthrough
					case INVENTORY_TRANSFER_STATUS_OK:
						*player.inbox <- fmt.Sprintf("You took %s from %s.", itemNameWithAmount(result.itemName, result.amount), chestName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED:
						*player.inbox <- "You must specify an item to take."
					case INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND:
						*player.inbox <- fmt.Sprintf("There is no item called '%s' in %s.", result.itemName, chestName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS:
						*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in %s.", result.itemName, chestName)
					case INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE:
						*player.inbox <- fmt.Sprintf("There is no item matching that number in %s.", chestName)
					case INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK:
						*player.inbox <- fmt.Sprintf("You can only take 1 %s at once.", result.itemName)
					default:
						panic(fmt.Sprintf("Transfer result status %d not handled.", result.status))
				}

				return true
			},
		},

		"loot": {
			usage: "loot <container>",
			description: "Take all items from the container in this room. If you specify 'room' as the container, you will take all items from the floor in this room.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Get pointer to room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				targetInventory, chestName, err := fuzzyFindChestInventory(playerRoom, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				if len(targetInventory.Items) == 0 {
					*player.inbox <- fmt.Sprintf("%s is empty.", chestName)
					return true
				}

				itemNames := make([]string, 0, len(targetInventory.Items))
				for len(targetInventory.Items) > 0 {
					item := targetInventory.RemoveStack(len(targetInventory.Items) - 1)

					playerMob.Data.Inventory.AddItem(item)
					itemNames = append(itemNames, item.GetNameWithAmount())
				}
				*player.inbox <- fmt.Sprintf("You got %s.", combineNames(itemNames))

				return true
			},
		},

		"craft": {
			usage: "craft <item>",
			description: "Craft an item for which you know the recipe",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				recipe, err := fuzzyFindKnownRecipe(player.character, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				recipeData := world.RECIPE_DATA[recipe]

				// Check if the player has the materials
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				for _, ingredient := range recipeData.Materials {
					amountOfIngredient := playerMob.Data.Inventory.AmountOf(ingredient.Id)
					if amountOfIngredient < ingredient.Amount {
						*player.inbox <- fmt.Sprintf("You lack the ingredients to craft %s", recipeData.Name)
						return true
					}
				}

				// Remove the materials from the player's inventory
				for _, ingredient := range recipeData.Materials {
					amountToRemove := ingredient.Amount
					for amountToRemove > 0 {
						index, _ := playerMob.Data.Inventory.FindItem(ingredient.Id)
						item := playerMob.Data.Inventory.RemoveItems(index, amountToRemove)
						amountToRemove -= item.Amount
					}
				}

				// Add the crafted item to the player's inventory
				recipeOutput := recipeData.CreateOutput()
				playerMob.Data.Inventory.AddItem(recipeOutput)
				*player.inbox <- fmt.Sprintf("You crafted %s.", recipeOutput.GetNameWithAmount())

				return true
			},
		},

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
					recipeList := "You know the following recipes:\n"
					for _, name := range player.character.RecipesKnown {
						recipeList = (recipeList + world.RECIPE_DATA[name].Name + "\n")
					}
					*player.inbox <- recipeList
					return true
				}

				if args[0] == "info" {
					query, err := fuzzyFindKnownRecipe(player.character, args[1:])

					if err != nil {
						*player.inbox <- err.Error()
						return true
					}

					recipeData := world.RECIPE_DATA[query]
					*player.inbox <- "The following recipe requires the following ingredients:"
					for _, ingredient := range recipeData.Materials {
						material := world.ITEM_DATA[ingredient.Id].Name
						possessed := playerMob.Data.Inventory.AmountOf(ingredient.Id)
						needed := ingredient.Amount
						*player.inbox <- fmt.Sprintf("%s: %d / %d", material, possessed, needed)
					}

					return true
				}

				return false
			},
		},

		"equipment": {
			usage: "equipment",
			description: "Show your current equipment",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				*player.inbox <- "Your equipment is:"
				printMobEquipmentList(player, playerMob)

				return true
			},
		},

		"equip": {
			usage: "equip <item> [in <slot>]",
			description: "Equip the specified item [in the specified slot].",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 1 {
					return false
				}

				itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "in")
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				// Determine the item
				itemIndex := fuzzyFindInventoryItemIndex(&playerMob.Data.Inventory, itemWords)

				// Handle edge cases
				if itemIndex == FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED {
					*player.inbox <- "You must specify an item to equip."
					return true
				}
				if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
					*player.inbox <- fmt.Sprintf("You have no item called '%s' in your inventory.",
						strings.Join(itemWords, " "))
					return true
				}
				if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
					*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in your inventory.",
						strings.Join(itemWords, " "))
					return true
				}

				// Check stat requirements
				item := &playerMob.Data.Inventory.Items[itemIndex]
				itemData := world.ITEM_DATA[item.Id]
				if !playerMob.Data.Stats.Meets(item.GetStatRequirements()) {
					*player.inbox <- fmt.Sprintf("You do not meet the stat requirements to equip %s.", item.GetNameWithCondition())
					return true
				}

				// Determine the equipment slot
				var slot world.EquipmentSlot
				if userSpecifiedSlot {
					if len(slotWords) == 0 {
						*player.inbox <- "When specifying 'in' you must also specify a slot."
						return false
					}

					var err error
					slot, err = fuzzyFindEquipmentSlot(slotWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				} else {
					if itemData.ItemIsOneHanded() {
						*player.inbox <- fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'main hand' hand or 'off hand'.", item.GetNameWithCondition())
						return false
					}

					var slotFound bool
					slot, slotFound = world.EquipmentSlotForItemType(itemData.ItemType)
					if !slotFound {
						*player.inbox <- fmt.Sprintf("%s cannot be equipped.", item.GetNameWithCondition())
						return true
					}
				}

				// Try to equip item
				unequippedItems, success := playerMob.Data.Equipment.Equip(slot, *item)
				if !success {
					*player.inbox <- fmt.Sprintf("%s cannot be equipped to slot %s.", item.GetNameWithCondition(), world.EquipmentSlotToString(slot))
					return true
				}
				*player.inbox <- fmt.Sprintf("You equipped %s.", item.GetNameWithCondition())

				// Remove item from player inventory
				playerMob.Data.Inventory.RemoveItem(itemIndex)

				// Add unequipped items to inventory
				for _, unequippedItem := range unequippedItems {
					message := playerMob.OnPlayerItemUnequipped(unequippedItem)
					if message != "" {
						*player.inbox <- fmt.Sprintf("%s was unequipped and added to your inventory", unequippedItem.GetNameWithCondition())
					}
				}

				// If the equipped item is a spellbook, add the spell to their spells equipped
				if (itemData.ItemType == world.ITEM_TYPE_EQUIPMENT_SPELLBOOK) {
					spellbookData := itemData.Data.(*world.ItemDataSpellbook)

					// Increment spell equiped count
					_, entryExists := player.character.SpellsEquipped[spellbookData.Spell]
					if !entryExists {
						player.character.SpellsEquipped[spellbookData.Spell] = &world.CharacterEquippedSpell {
							EquipCount: 0,
							Casts: 0,
							IsKnown: slices.Contains(player.character.SpellsKnown, spellbookData.Spell),
						}
					}
					player.character.SpellsEquipped[spellbookData.Spell].EquipCount++

					isSpellKnown := slices.Contains(player.character.SpellsKnown, spellbookData.Spell)
					if !isSpellKnown {
						spellData := world.SPELL_DATA[spellbookData.Spell]
						*player.inbox <- fmt.Sprintf("You can now prepare the spell %s.", spellData.Name)
					}
				}

				return true
			},
		},

		"remove": {
			usage: "remove [<item>] [from <slot>]",
			description: "Remove an equipped item by name or by slot.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// The way this works is that the user can specify a slot OR an item,
				// but there's no reason for them to specify both so I'm not going to
				// bother writing the code for it

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "from")

				var slot world.EquipmentSlot
				if userSpecifiedSlot {
					if len(itemWords) != 0 {
						*player.inbox <- "When specifying 'from', you should not specify an item."
						return false
					}
					if len(slotWords) == 0 {
						*player.inbox <- "When specifying 'from', you must also specify a slot."
						return false
					}

					// Find slot
					var err error
					slot, err = fuzzyFindEquipmentSlot(slotWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				} else {
					// Find slot
					var err error
					slot, err = fuzzyFindEquipmentSlotByItem(&playerMob.Data.Equipment, itemWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				}

				// Unequip the item
				item, _ := playerMob.Data.Equipment.Unequip(slot)
				message := playerMob.OnPlayerItemUnequipped(item)
				playerMob.Data.Inventory.AddItem(item)
				if message != "" {
					*player.inbox <- message
				}
				*player.inbox <- fmt.Sprintf("You unequipped %s.", item.GetNameWithCondition())

				return true
			},
		},

		"spells": {
			usage: "spells",
			description: "Show a list of the spells you have prepared",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				*player.inbox <- fmt.Sprintf("Spells Prepared (%d / %d):", len(playerMob.Data.Spells), playerMob.Data.SpellSlots())

				if len(playerMob.Data.Spells) == 0 {
					*player.inbox <- "You haven't prepared any spells."
					return true
				}

				for _, spell := range playerMob.Data.Spells {
					spellData := world.SPELL_DATA[spell]

					masteryStr := "Known"

					equippedSpell, spellIsEquipped := player.character.SpellsEquipped[spell]
					if spellIsEquipped && !equippedSpell.IsKnown {
						casts := float32(equippedSpell.Casts)
						castsToLearn := float32(playerMob.Data.CastsToLearn(spell))
						mastery := int32((casts / castsToLearn) * 100.0)
						masteryStr = fmt.Sprintf("Mastery: %d", mastery)
					}

					*player.inbox <- fmt.Sprintf("%s | Cost: %d | %s | %s",
						spellData.Name, spellData.ManaCost, masteryStr, spellData.Description)
				}

				return true
			},
		},

		"library": {
			usage: "library",
			description: "Show a list of all the spells you know",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(player.character.SpellsKnown) == 0 {
					*player.inbox <- "You don't know any spells."
				}

				for _, spell := range player.character.SpellsKnown {
					spellData := world.SPELL_DATA[spell]
					*player.inbox <- fmt.Sprintf("%s - Cost: %d - %s",
						spellData.Name, spellData.ManaCost, spellData.Description)
				}

				return true
			},
		},

		"prepare": {
			usage: "prepare <spell>",
			description: "Prepare a spell from the list of spells you know",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Make sure the player is in a safe zone
				if !playerRoom.IsSafeZone {
					*player.inbox <- "You can only prepare spells from within a safe room."
					return true
				}

				// Check if there is an empty spell slot
				if len(playerMob.Data.Spells) >= int(playerMob.Data.SpellSlots()) {
					*player.inbox <- "You don't have any available spell slots."
					*player.inbox <- "Type 'forget <spell>' to free up a spell slot."
					return true
				}

				// Check if the spell is already prepared
				spell, err := fuzzyFindPreparedSpell(gamestate, player, args)
				isPrepared := err == nil
				if isPrepared {
					spellData := world.SPELL_DATA[spell]
					*player.inbox <- fmt.Sprintf("You have already prepared %s.", spellData.Name)
					return true
				}

				// Search for spell
				spell, err = fuzzyFindKnownOrEquippedSpell(player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				playerMob.Data.Spells = append(playerMob.Data.Spells, spell)
				*player.inbox <- fmt.Sprintf("You prepared %s.", world.SPELL_DATA[spell].Name)
				return true
			},
		},

		"forget": {
			usage: "forget <spell>",
			description: "Removes a spell from your prepared spells list. If you have mastered the spell, it will remain in your library.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 1 {
					return false
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Make sure the player is in a safe zone
				if !playerRoom.IsSafeZone {
					*player.inbox <- "You can only forget spells from within a safe room."
					return true
				}

				// Find a spell that matches their input and remove it
				spell, err := fuzzyFindPreparedSpell(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				playerMob.Data.RemoveSpell(spell)
				*player.inbox <- fmt.Sprintf("You forgot %s.", world.SPELL_DATA[spell].Name)
				return true
			},
		},

		"cast": {
			usage: "cast <spell> at <target>",
			description: "Cast a spell",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				spellWords, targetWords, userSpecifiedTarget := splitArgsBy(args, "at")

				if len(spellWords) == 0 || len(targetWords) == 0 || !userSpecifiedTarget {
					return false
				}

				// Find the spell in their spell list
				spell, err := fuzzyFindPreparedSpell(gamestate, player, spellWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Find the target in the room
				targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Check against PvP
				spellData := world.SPELL_DATA[spell]
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter != nil && !spellData.CanTargetPlayers {
					*player.inbox <- "You cannot cast that spell against players."
					return true
				}

				// Queue up a cast action
				player.nextAction = Action {
					actionType: ACTION_TYPE_CAST,
					data: ActionCast {
						spell: spell,
						target: targetHandle,
					},
				}

				return true
			},
		},

		"use": {
			usage: "use <item> [on <target>]",
			description: "Use an item. If you don't specify a target, the target is assumed to be yourself.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				itemWords, targetWords, userSpecifiedTarget := splitArgsBy(args, "on")

				// For now, all consumables can only be used on "self"
				// Spell scrolls can be used on others based on the spell's targeting rules

				if len(itemWords) == 0 {
					return false
				}
				if userSpecifiedTarget && len(targetWords) == 0 {
					*player.inbox <- "When specifying 'on' you must specify a target."
					return false
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				itemIndex := fuzzyFindInventoryItemIndex(&playerMob.Data.Inventory, itemWords)

				// Handle edge cases
				if itemIndex == FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED {
					*player.inbox <- "You must specify an item to use."
					return true
				}
				if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
					*player.inbox <- fmt.Sprintf("You have no item called '%s' in your inventory.",
						strings.Join(itemWords, " "))
					return true
				}
				if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
					*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in your inventory.",
						strings.Join(itemWords, " "))
					return true
				}

				// Determine target
				var targetHandle world.MobHandle
				if userSpecifiedTarget {
					var err error
					targetHandle, err = fuzzyFindTarget(gamestate, player, targetWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				} else {
					targetHandle = player.mobHandle
				}

				// Check for item <-> target compatibility
				item := &playerMob.Data.Inventory.Items[itemIndex]
				itemData := world.ITEM_DATA[item.Id]

				switch itemData.ItemType {
					case world.ITEM_TYPE_CONSUMABLE:
						if targetHandle != player.mobHandle {
							*player.inbox <- "That item can only be used on yourself."
							return true
						}

						player.nextAction = Action {
							actionType: ACTION_TYPE_USE_ITEM,
							data: ActionUseItem {
								itemId: item.Id,
								target: targetHandle,
							},
						}

						return true
					case world.ITEM_TYPE_SPELL_SCROLL:
						scrollData := itemData.Data.(*world.ItemDataSpellScroll)
						spellInfo := world.SPELL_DATA[scrollData.Spell]

						targetMob := gamestate.world.Mobs.Get(targetHandle)
						if targetMob.PlayerCharacter != nil && !spellInfo.CanTargetPlayers {
							*player.inbox <- "You cannot cast that spell against players."
							return true
						}

						player.nextAction = Action {
							actionType: ACTION_TYPE_USE_ITEM,
							data: ActionUseItem {
								itemId: item.Id,
								target: targetHandle,
							},
						}

						return true
					case world.ITEM_TYPE_RECIPE:
						var recipe world.Recipe = itemData.Data.(*world.ItemDataRecipe).Recipe
						recipeData := world.RECIPE_DATA[recipe]

						if recipeData.Job != player.character.Job {
							*player.inbox <- fmt.Sprintf("You must be a %s to learn that recipe.", world.JOB_DATA[recipeData.Job].Name)
							return true
						}

						if playerMob.Data.Level < recipeData.Level {
							*player.inbox <- "Your level is not high enough to learn this recipe."
							return true
						}

						if slices.Contains(player.character.RecipesKnown, recipe) {
							*player.inbox <- "You already know this recipe."
							return true
						}

						player.character.RecipesKnown = append(player.character.RecipesKnown, recipe)
						*player.inbox <- fmt.Sprintf("You have learned the recipe for %s.", recipeData.Name)

						playerMob.Data.Inventory.RemoveItem(itemIndex)
						return true
					default:
						*player.inbox <- "That item is not a consumable."
						return true
				}
			},
		},

		"trade": {
			usage: "trade <action>",
			description: "Trade atomically with another player. Type 'trade help' to see a list of trade actions.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				action := strings.ToLower(args[0])
				if action == "help" {
					// User asked for help about a specific command
					if len(args) >= 2 {
						// Lookup the command in the menu
						entry, entryExists := MENU_TRADE_ENTRIES[strings.ToLower(args[1])]
						if !entryExists {
							*player.inbox <- fmt.Sprintf("Cannot provide help because %s is not a known trading command.", args[1])
							return true
						}

						// Send the help info to the user
						*player.inbox <- fmt.Sprintf("%s - %s", entry.usage, entry.description)
						return true
					}

					// Provide help info
					*player.inbox <- "The trade menu lets you trade with another player. You can only trade with one player at a time."
					*player.inbox <- "Commands:"
					for _, entry := range MENU_TRADE_ENTRIES {
						*player.inbox <- fmt.Sprintf("\t%s - %s", entry.usage, entry.description)
					}

					return true
				}

				entry, entryExists := MENU_TRADE_ENTRIES[action]
				if !entryExists {
					*player.inbox <- fmt.Sprintf("%s is not a trade action. Type 'trade help' to see a list of actions.", action)
					return true
				}

				success := entry.handler(gamestate, player, args[1:])
				if !success {
					*player.inbox <- fmt.Sprintf("Invalid command. Usage: %s", entry.usage)
				}

				return true
			},
		},

		"rest": {
			usage: "rest",
			description: "Take a rest to regain your health and mana",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Make sure the player is in a safe zone
				if !playerRoom.IsSafeZone {
					*player.inbox <- "You can only rest from within a safe room."
					return true
				}

				gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s took a nap.", playerMob.Data.Name))
				*player.inbox <- "Your HP and MP have been restored!"

				playerMob.Data.Health = playerMob.Data.MaxHealth()
				playerMob.Data.Mana = playerMob.Data.MaxMana()

				return true
			},
		},
	},
}

func describeRoomToPlayer(gamestate *GameState, player *Player, room *world.Room) {
	*player.inbox <- room.Description

	// Chests
	if len(room.Chests) > 0 {
		chestNames := make([]string, 0, len(room.Chests))
		for index := range len(room.Chests) {
			chest := &room.Chests[index]
			chestNames = append(chestNames, chest.Name)
		}

		*player.inbox <- fmt.Sprintf("In this room is %s", combineNames(chestNames))
	}

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

		*player.inbox <- fmt.Sprintf("To the %s is %s.", world.DirectionToString(direction), roomName)
		exitFound = true
	}
	if !exitFound {
		*player.inbox <- "This room has no exits!"
	}

	// Send the list of players in the room
	if len(room.Occupants) > 1 {
		otherPlayerCount := len(room.Occupants) - 1

		otherPlayerNames := make([]string, 0, otherPlayerCount)
		for _, mobHandle := range room.Occupants {
			// Don't tell the player about themselves being in the room
			if mobHandle == player.mobHandle {
				continue
			}

			// Get a pointer to the mob
			mob := gamestate.world.Mobs.Get(mobHandle)
			// Add their name to the list
			otherPlayerNames = append(otherPlayerNames, mob.GetName())
		}

		otherPlayersStr := combineNames(otherPlayerNames)
		isString := "are"
		if otherPlayerCount == 1 {
			isString = "is"
		}
		*player.inbox <- fmt.Sprintf("%s %s here.", otherPlayersStr, isString)
	}
}

func printMobEquipmentList(player *Player, mob *world.Mob) {
	// Determine if we should skip the offhand item slot
	mainHandItem := mob.Data.Equipment.Get(world.EQUIPMENT_SLOT_MAIN_HAND)
	shouldSkipOffhand := mainHandItem != nil && world.ITEM_DATA[mainHandItem.Id].ItemType == world.EQUIPMENT_SLOT_MAIN_HAND

	for index := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(index)
		item := mob.Data.Equipment.Get(slot)
		var itemData *world.ItemData = nil

		// If two handed equipped, skip off hand
		if slot == world.EQUIPMENT_SLOT_OFF_HAND && shouldSkipOffhand {
			continue
		}

		// Determine item name
		var itemName string
		if item != nil {
			itemData = world.ITEM_DATA[item.Id]
			itemName = item.GetNameWithCondition()
		} else {
			itemName = "<Nothing Equipped>"
		}

		// Determine slot name
		var slotName string
		if slot == world.EQUIPMENT_SLOT_MAIN_HAND && item != nil && itemData.ItemType == world.ITEM_TYPE_EQUIPMENT_TWO_HANDED {
			slotName = "Both Hands"
		} else {
			slotName = world.EquipmentSlotToString(slot)
		}

		*player.inbox <- fmt.Sprintf("\t%s - %s", slotName, itemName)
	}
}

func printMobHp(player *Player, mob *world.Mob) {
	*player.inbox <- fmt.Sprintf("%s:", mob.Data.Name)
	*player.inbox <- fmt.Sprintf("HP: %d / %d", mob.Data.Health, mob.Data.MaxHealth())
	*player.inbox <- fmt.Sprintf("MP: %d / %d", mob.Data.Mana, mob.Data.MaxMana())
}
