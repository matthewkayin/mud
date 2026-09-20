package game

import (
	"fmt"
	"mud/bitset"
	"slices"
	"strconv"
	"strings"
)

type InventoryFindResult int
const (
	INVENTORY_FIND_RESULT_NOT_FOUND = iota
	INVENTORY_FIND_RESULT_AMBIGUOUS
	INVENTORY_FIND_RESULT_FOUND
)

const INVENTORY_TRANSFER_AMOUNT_ALL = -1

type InventoryTransferStatus int
const (
	INVENTORY_TRANSFER_STATUS_OK = iota
	INVENTORY_TRANSFER_STATUS_PARTIAL
	INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED
	INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND
	INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS
	INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE
	INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK
)

type InventoryTransferResult struct {
	status InventoryTransferStatus
	amount int32
	itemName string
}

func MenuWorld() Menu {
	entries := make(map[string]MenuEntry)

	// Logout
	entries["logout"] = MenuEntry {
		usage: "logout",
		description: "Logout of the world",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			player.exitWorld(gameState)
			return true
		},
	}

	// Who
	entries["who"] = MenuEntry {
		usage: "who",
		description: "Get a list of everyone who is online",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			names := make([]string, 0, len(gameState.players))
			for index := range len(gameState.players) {
				if !gameState.players[index].isLoggedIn {
					continue
				}

				names = append(names, gameState.players[index].character.Data.Name)
			}

			*player.inbox <- fmt.Sprintf("The players in the world are: %s.", combineNames(names))

			return true
		},
	}

	// Look
	entries["look"] = MenuEntry {
		usage: "look",
		description: "Describe the current room",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.data.Room]

			describeRoomToPlayer(gameState, player, room)
			return true
		},
	}

	// Search
	entries["search"] = MenuEntry {
		usage: "search [container]",
		description: "Search the room or a container for items.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.data.Room]

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
				itemNames = append(itemNames, targetInventory.Items[index].getNameWithAmount())
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
	}

	// Exits
	entries["exits"] = MenuEntry {
		usage: "exits",
		description: "Describe the exits of the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.data.Room]

			exitFound := false
			for directionIndex := range DIRECTION_COUNT {
				direction := Direction(directionIndex)

				if room.Exits[direction] == ROOM_NONE {
					continue
				}

				roomName := "an undiscovered room"
				if bitset.Check(player.character.RoomsDiscovered, room.Exits[direction]) {
					roomName = gameState.world.Rooms[room.Exits[direction]].Name
				}

				*player.inbox <- fmt.Sprintf("To the %s is %s.", DirectionToString(direction), roomName)
				exitFound = true
			}

			if !exitFound {
				*player.inbox <- "This room has no exits!"
			}

			return true
		},
	}

	// Say
	entries["say"] = MenuEntry {
		usage: "say <message>",
		description: "Send a messsage to the current room",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]
			playerRoom.broadcast(gameState, fmt.Sprintf("%s said '%s'",
				player.character.Data.Name,
				strings.TrimSpace(strings.Join(args, " "))))
			return true
		},
	}

	// Tell
	entries["tell"] = MenuEntry {
		usage: "tell <target> <message>",
		description: "Send a message to a specific person in the current room",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// Parse input
			argString := strings.Join(args, " ")
			targetString, messageString, colonFound := strings.Cut(argString, ":")
			message := strings.TrimSpace(messageString)
			if !colonFound || len(targetString) == 0 || len(message) == 0 {
				return false
			}

			// Find target
			targetHandle, err := fuzzyFindTarget(gameState, player, strings.Fields(targetString))
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
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			targetMob := gameState.world.Mobs.Get(targetHandle)
			if targetMob.player != nil {
				*player.inbox <- fmt.Sprintf("You told %s: '%s'", playerMob.data.Name, message)
				*targetMob.player.inbox <- fmt.Sprintf("%s told you: '%s'", playerMob.data.Name, message)
				return true
			}

			// TODO: handle case where they talk to an NPC

			return true
		},
	}

	// Yell
	entries["yell"] = MenuEntry {
		usage: "yell <message>",
		description: "Send a message to everyone in this and the adjacent rooms",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Get a handle to the player room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

			// Collect a list containing all rooms to broadcast to
			rooms := make([]*Room, 0, 5)
			rooms = append(rooms, playerRoom)

			// Add room exits to the list
			for direction := range DIRECTION_COUNT {
				roomIndex := playerRoom.Exits[direction]
				if roomIndex == ROOM_NONE {
					continue
				}

				adjacentRoom := &gameState.world.Rooms[roomIndex]
				rooms = append(rooms, adjacentRoom)
			}

			// Broadcast the message to each room
			message := fmt.Sprintf("%s: '%s'",
				playerMob.data.Name,
				strings.TrimSpace(strings.Join(args, " ")))
			for _, room := range rooms {
				room.broadcast(gameState, message)
			}

			return true
		},
	}

	// Move
	entries["move"] = MenuEntry {
		usage: "move <direction>",
		description: "Walk to an adjacent room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// Get the player mob and room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

			// Determine the index of the target room
			direction, directionFound := DirectionFromString(args[0])
			if !directionFound {
				*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
				return true
			}

			// Move player
			err := playerRoom.MoveOccupant(gameState, player.mobHandle, direction)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			playerRoom = &gameState.world.Rooms[playerMob.data.Room]
			*player.inbox <- fmt.Sprintf("You moved into %s.", playerRoom.Name)
			describeRoomToPlayer(gameState, player, playerRoom)
			return true
		},
	}

	// HP
	entries["hp"] = MenuEntry {
		usage: "hp",
		description: "Show your combat status including HP, MP, and conditions.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			*player.inbox <- fmt.Sprintf("HP: %d / %d", playerMob.data.Health, playerMob.data.MaxHealth())
			*player.inbox <- fmt.Sprintf("MP: %d / %d", playerMob.data.Mana, playerMob.data.MaxMana())
			return true
		},
	}

	// Stats
	entries["stats"] = MenuEntry {
		usage: "stats",
		description: "Show your current stats.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			classData := CLASS_DATA[player.character.Class]
			raceData := RACE_DATA[player.character.Race]

			*player.inbox <- fmt.Sprintf("%s - Level %d %s %s",
				playerMob.data.Name, playerMob.data.Level, raceData.Name, classData.Name)
			*player.inbox <- fmt.Sprintf("Experience: %d / %d", playerMob.data.Experience, playerMob.data.ExperienceToNextLevel)

			*player.inbox <- fmt.Sprintf("\nHP: %d / %d", playerMob.data.Health, playerMob.data.MaxHealth())
			*player.inbox <- fmt.Sprintf("MP: %d / %d", playerMob.data.Mana, playerMob.data.MaxMana())

			statBonuses := &playerMob.data.EquippedItems.statBonuses
			*player.inbox <- fmt.Sprintf("\nVitality: %d (+%d)", playerMob.data.Stats.Vitality, statBonuses.Vitality)
			*player.inbox <- fmt.Sprintf("Strength: %d (+%d)", playerMob.data.Stats.Strength, statBonuses.Strength)
			*player.inbox <- fmt.Sprintf("Agility: %d (+%d)", playerMob.data.Stats.Agility, statBonuses.Agility)
			*player.inbox <- fmt.Sprintf("Intelligence: %d (+%d)", playerMob.data.Stats.Intelligence, statBonuses.Intelligence)
			*player.inbox <- fmt.Sprintf("Faith: %d (+%d)", playerMob.data.Stats.Faith, statBonuses.Faith)

			return true
		},
	}

	// Stop
	entries["stop"] = MenuEntry {
		usage: "stop",
		description: "Stop attacking or cancel your current spell",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

			switch playerMob.mode {
				case MOB_MODE_IDLE:
				case MOB_MODE_ATTACK:
					targetMob, targetExists := gameState.world.Mobs.GetIfExists(playerMob.target)
					if targetExists {
						playerRoom.broadcast(gameState, fmt.Sprintf("%s stopped attacking %s", playerMob.data.Name, targetMob.data.Name))
					}
				case MOB_MODE_CAST:
					playerRoom.broadcast(gameState, fmt.Sprintf("%s canceled their spell.", playerMob.data.Name))
			}

			playerMob.mode = MOB_MODE_IDLE
			player.nextAction = Action {
				actionType: ACTION_TYPE_NONE,
				data: nil,
			}

			return true
		},
	}

	// Attack
	entries["attack"] = MenuEntry {
		usage: "attack <target>",
		description: "Attack the target.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			targetHandle, err := fuzzyFindTarget(gameState, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check for PvP
			targetMob := gameState.world.Mobs.Get(targetHandle)
			if targetMob.player != nil {
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
	}

	// Inventory
	entries["inventory"] = MenuEntry {
		usage: "inventory",
		description: "List the items in your inventory",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			inventorySize := len(playerMob.data.Inventory.Items)

			if playerMob.data.Inventory.Length() == 0 {
				*player.inbox <- "There is nothing in your inventory."
				return true
			}

			for index := range inventorySize {
				item := &playerMob.data.Inventory.Items[index]
				itemData := ITEM_DATA[item.Id]
				itemStatRequirements := item.getStatRequirements()

				typeStr := ItemTypeToString(itemData.itemType)
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

				itemName := item.getNameWithCondition()
				if item.Amount > 1 {
					itemName = fmt.Sprintf("%s (x%d)", itemName, item.Amount)
				}
				*player.inbox <- fmt.Sprintf("%s | Type: %s | Description: %s", itemName, typeStr, itemData.description)
			}
			itemNames := make([]string, 0, inventorySize)
			for _, item := range playerMob.data.Inventory.Items {
				itemNames = append(itemNames, ITEM_DATA[item.Id].name)
			}
			return true
		},
	}

	// Drop an item
	entries["drop"] = MenuEntry {
		usage: "drop <item>",
		description: "Drop an item from your inventory",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

			result := inventoryTransfer(&playerMob.data.Inventory, &playerRoom.Inventory, args)
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
	}

	// Put an item into a container
	entries["put"] = MenuEntry {
		usage: "put <item> into <container>",
		description: "Puts an item from your inventory into the specified container",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.data.Room]

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
			result := inventoryTransfer(&playerMob.data.Inventory, targetInventory, itemWords)
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
	}

	// Give an item
	entries["give"] = MenuEntry {
		usage: "give <item> to <target>",
		description: "Gives an item to a player",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			// Split args
			itemWords, targetWords, userSpecifiedTo := splitArgsBy(args, "to")
			if !userSpecifiedTo || len(itemWords) == 0 || len(targetWords) == 0 {
				return false
			}

			// Find target
			targetHandle, err := fuzzyFindTarget(gameState, player, targetWords)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// TODO: allow giving to non-player NPCs for things like RP and encounters?
			targetMob := gameState.world.Mobs.Get(targetHandle)
			if targetMob.player == nil {
				*player.inbox <- "You cannot give an item to someone who isn't a player."
				return true
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			result := inventoryTransfer(&playerMob.data.Inventory, &targetMob.data.Inventory, itemWords)
			switch result.status {
				case INVENTORY_TRANSFER_STATUS_PARTIAL:
					*player.inbox <- fmt.Sprintf("You only have %d %s in your inventory.", result.amount, result.itemName)
					fallthrough
				case INVENTORY_TRANSFER_STATUS_OK:
					*player.inbox <- fmt.Sprintf("You gave %s to %s.", itemNameWithAmount(result.itemName, result.amount), targetMob.data.Name)
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
	}

	// Grab an item
	entries["take"] = MenuEntry {
		usage: "take <item> [from <container>]",
		description: "Pick up an item in your current room",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			itemWords, chestWords, userSpecifiedChest := splitArgsBy(args, "from")
			if userSpecifiedChest && len(chestWords) == 0 {
				return false
			}

			// Get pointer to room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

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
			result := inventoryTransfer(targetInventory, &playerMob.data.Inventory, itemWords)
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
	}

	// Loot items
	entries["loot"] = MenuEntry {
		usage: "loot <container>",
		description: "Take all items from the container in this room. If you specify 'room' as the container, you will take all items from the floor in this room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// Get pointer to room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

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

				playerMob.data.Inventory.AddItem(item)
				itemNames = append(itemNames, item.getNameWithAmount())
			}
			*player.inbox <- fmt.Sprintf("You got %s.", combineNames(itemNames))

			return true
		},
	}

	// Show equipment
	entries["equipment"] = MenuEntry {
		usage: "equipment",
		description: "Show your current equipment",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Determine if we should skip the offhand item slot
			mainHandItem := playerMob.data.EquippedItems.Get(EQUIPMENT_SLOT_MAIN_HAND)
			shouldSkipOffhand := mainHandItem != nil && ITEM_DATA[mainHandItem.Id].itemType == EQUIPMENT_SLOT_MAIN_HAND

			*player.inbox <- "Your equipment is:"

			for index := range EQUIPMENT_SLOT_COUNT {
				slot := EquipmentSlot(index)
				item := playerMob.data.EquippedItems.Get(slot)
				var itemData *ItemData = nil

				// If two handed equipped, skip off hand
				if slot == EQUIPMENT_SLOT_OFF_HAND && shouldSkipOffhand {
					continue
				}

				// Determine item name
				var itemName string
				if item != nil {
					itemData = ITEM_DATA[item.Id]
					itemName = item.getNameWithCondition()
				} else {
					itemName = "<Nothing Equipped>"
				}

				// Determine slot name
				var slotName string
				if slot == EQUIPMENT_SLOT_MAIN_HAND && item != nil && itemData.itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED {
					slotName = "Both Hands"
				} else {
					slotName = EquipmentSlotToString(slot)
				}

				*player.inbox <- fmt.Sprintf("\t%s - %s", slotName, itemName)
			}

			return true
		},
	}

	// Equip
	entries["equip"] = MenuEntry {
		usage: "equip <item> [in <slot>]",
		description: "Equip the specified item [in the specified slot].",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "in")
			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Determine the item
			itemIndex := fuzzyFindInventoryItemIndex(&playerMob.data.Inventory, itemWords)

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
			item := &playerMob.data.Inventory.Items[itemIndex]
			itemData := ITEM_DATA[item.Id]
			if !playerMob.data.Stats.Meets(item.getStatRequirements()) {
				*player.inbox <- fmt.Sprintf("You do not meet the stat requirements to equip %s.", item.getNameWithCondition())
				return true
			}

			// Determine the equipment slot
			var slot EquipmentSlot
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
					*player.inbox <- fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'main hand' hand or 'off hand'.", item.getNameWithCondition())
					return false
				}

				var slotFound bool
				slot, slotFound = EquipmentSlotForItemType(itemData.itemType)
				if !slotFound {
					*player.inbox <- fmt.Sprintf("%s cannot be equipped.", item.getNameWithCondition())
					return true
				}
			}

			// Try to equip item
			unequippedItems, success := playerMob.data.EquippedItems.Equip(slot, *item)
			if !success {
				*player.inbox <- fmt.Sprintf("%s cannot be equipped to slot %s.", item.getNameWithCondition(), EquipmentSlotToString(slot))
				return true
			}
			*player.inbox <- fmt.Sprintf("You equipped %s.", item.getNameWithCondition())

			// Remove item from player inventory
			playerMob.data.Inventory.RemoveItem(itemIndex)

			// Add unequipped items to inventory
			for _, unequippedItem := range unequippedItems {
				player.onItemUnequipped(gameState, unequippedItem)
				*player.inbox <- fmt.Sprintf("%s was unequipped and added to your inventory", unequippedItem.getNameWithCondition())
			}

			// If the equipped item is a spellbook, add the spell to their spells equipped
			if (itemData.itemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK) {
				spellbookData := itemData.data.(*ItemDataSpellbook)

				// Increment spell equiped count
				_, entryExists := player.character.SpellsEquipped[spellbookData.spell]
				if !entryExists {
					player.character.SpellsEquipped[spellbookData.spell] = &CharacterEquippedSpell {
						EquipCount: 0,
						Casts: 0,
						IsKnown: slices.Contains(player.character.SpellsKnown, spellbookData.spell),
					}
				}
				player.character.SpellsEquipped[spellbookData.spell].EquipCount++

				isSpellKnown := slices.Contains(player.character.SpellsKnown, spellbookData.spell)
				if !isSpellKnown {
					spellData := SPELL_DATA[spellbookData.spell]
					*player.inbox <- fmt.Sprintf("You can now prepare the spell %s.", spellData.name)
				}
			}

			return true
		},
	}

	// Unequip
	entries["remove"] = MenuEntry {
		usage: "remove [<item>] [from <slot>]",
		description: "Remove an equipped item by name or by slot.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// The way this works is that the user can specify a slot OR an item,
			// but there's no reason for them to specify both so I'm not going to
			// bother writing the code for it

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "from")

			var slot EquipmentSlot
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
				slot, err = fuzzyFindEquipmentSlotByItem(&playerMob.data.EquippedItems, itemWords)
				if err != nil {
					return true
				}
			}

			// Unequip the item
			item, _ := playerMob.data.EquippedItems.Unequip(slot)
			player.onItemUnequipped(gameState, item)
			*player.inbox <- fmt.Sprintf("You unequipped %s.", item.getNameWithCondition())

			return true
		},
	}

	// Spells prepared
	entries["spells"] = MenuEntry {
		usage: "spells",
		description: "Show a list of the spells you have prepared",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			*player.inbox <- fmt.Sprintf("Spells Prepared (%d / %d):", len(playerMob.data.Spells), playerMob.data.SpellSlots())

			if len(playerMob.data.Spells) == 0 {
				*player.inbox <- "You haven't prepared any spells."
				return true
			}

			for _, spell := range playerMob.data.Spells {
				spellData := SPELL_DATA[spell]

				masteryStr := "Known"

				equippedSpell, spellIsEquipped := player.character.SpellsEquipped[spell]
				if spellIsEquipped && !equippedSpell.IsKnown {
					casts := float32(equippedSpell.Casts)
					castsToLearn := float32(playerMob.data.CastsToLearn(spell))
					mastery := int32((casts / castsToLearn) * 100.0)
					masteryStr = fmt.Sprintf("Mastery: %d", mastery)
				}

				*player.inbox <- fmt.Sprintf("%s | Cost: %d | %s | %s",
					spellData.name, spellData.manaCost, masteryStr, spellData.description)
			}

			return true
		},
	}

	// Spells known
	entries["library"] = MenuEntry {
		usage: "library",
		description: "Show a list of all the spells you know",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(player.character.SpellsKnown) == 0 {
				*player.inbox <- "You don't know any spells."
			}

			for _, spell := range player.character.SpellsKnown {
				spellData := SPELL_DATA[spell]
				*player.inbox <- fmt.Sprintf("%s - Cost: %d - %s",
					spellData.name, spellData.manaCost, spellData.description)
			}

			return true
		},
	}

	// Prepare
	entries["prepare"] = MenuEntry {
		usage: "prepare <spell>",
		description: "Prepare a spell from the list of spells you know",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// Check if there is an empty spell slot
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			if len(playerMob.data.Spells) >= int(playerMob.data.SpellSlots()) {
				*player.inbox <- "You don't have any available spell slots."
				*player.inbox <- "Type 'forget <spell>' to free up a spell slot."
				return true
			}

			// Check if the spell is already prepared
			spell, err := fuzzyFindPreparedSpell(gameState, player, args)
			isPrepared := err == nil
			if isPrepared {
				spellData := SPELL_DATA[spell]
				*player.inbox <- fmt.Sprintf("You have already prepared %s.", spellData.name)
				return true
			}

			// Search for spell
			spell, err = fuzzyFindKnownOrEquippedSpell(player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			playerMob.data.Spells = append(playerMob.data.Spells, spell)
			*player.inbox <- fmt.Sprintf("You prepared %s.", SPELL_DATA[spell].name)
			return true
		},
	}

	// Forget
	entries["forget"] = MenuEntry {
		usage: "forget <spell>",
		description: "Removes a spell from your prepared spells list. If you have mastered the spell, it will remain in your library.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// Find a spell that matches their input and remove it
			spell, err := fuzzyFindPreparedSpell(gameState, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.data.RemoveSpell(spell)
			*player.inbox <- fmt.Sprintf("You forgot %s.", SPELL_DATA[spell].name)
			return true
		},
	}

	// Cast
	entries["cast"] = MenuEntry {
		usage: "cast <spell> at <target>",
		description: "Cast a spell",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			spellWords, targetWords, userSpecifiedTarget := splitArgsBy(args, "at")

			if len(spellWords) == 0 || len(targetWords) == 0 || !userSpecifiedTarget {
				return false
			}

			// Find the spell in their spell list
			spell, err := fuzzyFindPreparedSpell(gameState, player, spellWords)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Find the target in the room
			targetHandle, err := fuzzyFindTarget(gameState, player, targetWords)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check against PvP
			spellData := SPELL_DATA[spell]
			targetMob := gameState.world.Mobs.Get(targetHandle)
			if targetMob.player != nil && !spellData.canTargetPlayers {
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
	}

	// Use
	entries["use"] = MenuEntry {
		usage: "use <item> [on <target>]",
		description: "Use an item. If you don't specify a target, the target is assumed to be yourself.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
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

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			itemIndex := fuzzyFindInventoryItemIndex(&playerMob.data.Inventory, itemWords)

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
			var targetHandle MobHandle
			if userSpecifiedTarget {
				var err error
				targetHandle, err = fuzzyFindTarget(gameState, player, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
			} else {
				targetHandle = player.mobHandle
			}

			// Check for item <-> target compatibility
			item := &playerMob.data.Inventory.Items[itemIndex]
			itemData := ITEM_DATA[item.Id]

			switch itemData.itemType {
				case ITEM_TYPE_CONSUMABLE:
					if targetHandle != player.mobHandle {
						*player.inbox <- "That item can only be used on yourself."
						return true
					}
				case ITEM_TYPE_SPELL_SCROLL:
					scrollData := itemData.data.(*ItemDataSpellScroll)
					spellInfo := SPELL_DATA[scrollData.spell]

					targetMob := gameState.world.Mobs.Get(targetHandle)
					if targetMob.player != nil && !spellInfo.canTargetPlayers {
						*player.inbox <- "You cannot cast that spell against players."
						return true
					}
				default:
					*player.inbox <- "That item is not a consumable."
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
		},
	}

	// Trade
	entries["trade"] = MenuEntry {
		usage: "trade <action>",
		description: "Trade atomically with another player. Type 'trade help' to see a list of trade actions.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
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

			success := entry.handler(gameState, player, args[1:])
			if !success {
				*player.inbox <- fmt.Sprintf("Invalid command. Usage: %s", entry.usage)
			}

			return true
		},
	}

	return Menu {
		entries: entries,
		getDescription: func (gameState *GameState, player *Player) string {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.data.Room]
			return fmt.Sprintf("You are in %s.", room.Name)
		},
	}
}

func describeRoomToPlayer(gameState *GameState, player *Player, room *Room) {
	*player.inbox <- room.Description

	// Send the list of players in the room
	if len(room.occupants) > 1 {
		otherPlayerCount := len(room.occupants) - 1

		otherPlayerNames := make([]string, 0, otherPlayerCount)
		for _, mobHandle := range room.occupants {
			// Don't tell the player about themselves being in the room
			if mobHandle == player.mobHandle {
				continue
			}

			// Get a pointer to the mob
			mob := gameState.world.Mobs.Get(mobHandle)
			// Add their name to the list
			otherPlayerNames = append(otherPlayerNames, mob.data.Name)
		}

		otherPlayersStr := combineNames(otherPlayerNames)
		isString := "are"
		if otherPlayerCount == 1 {
			isString = "is"
		}
		*player.inbox <- fmt.Sprintf("%s %s here.", otherPlayersStr, isString)
	}

	if len(room.Chests) > 0 {
		chestNames := make([]string, 0, len(room.Chests))
		for index := range len(room.Chests) {
			chest := &room.Chests[index]
			chestNames = append(chestNames, chest.Name)
		}

		*player.inbox <- fmt.Sprintf("In this room is %s", combineNames(chestNames))
	}
}

func inventoryTransfer(fromInventory *Inventory, toInventory *Inventory, itemWords []string) InventoryTransferResult {
	// Check for 0 item words
	if len(itemWords) == 0 {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED,
		}
	}

	// Get amount from args
	var amount int32 = 1
	if itemWords[0] == "all" {
		amount = INVENTORY_TRANSFER_AMOUNT_ALL
		itemWords = itemWords[1:]
	} else {
		parsedAmount, err := strconv.Atoi(itemWords[0])
		if err == nil {
			amount = int32(parsedAmount)
			itemWords = itemWords[1:]
		}
	}

	// Check for 0 item words once again
	if len(itemWords) == 0 {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED,
		}
	}

	// Find item
	itemIndex := fuzzyFindInventoryItemIndex(fromInventory, itemWords)

	// Handle edge cases on item index
	if itemIndex == FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED,
		}
	}
	if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS,
			itemName: strings.Join(itemWords, " "),
		}
	}
	if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND,
			itemName: strings.Join(itemWords, " "),
		}
	}
	if itemIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE,
			itemName: strings.Join(itemWords, " "),
		}
	}

	// Prevent user from transfering multiple of a non-stacking item
	itemData := ITEM_DATA[fromInventory.Items[itemIndex].Id]
	if amount != 1 && !itemData.ItemCanStack() {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK,
			itemName: fromInventory.Items[itemIndex].getNameWithCondition(),
		}
	}

	// Handle item amount "all"
	if amount == INVENTORY_TRANSFER_AMOUNT_ALL {
		amount = fromInventory.Items[itemIndex].Amount
	}

	// Transfer item
	removedItem := fromInventory.RemoveItems(itemIndex, amount)
	toInventory.AddItem(removedItem)

	// Determine result status
	resultStatus := INVENTORY_TRANSFER_STATUS_OK
	if removedItem.Amount < amount {
		resultStatus = INVENTORY_TRANSFER_STATUS_PARTIAL
	}

	return InventoryTransferResult {
		status: InventoryTransferStatus(resultStatus),
		amount: removedItem.Amount,
		itemName: removedItem.getNameWithCondition(),
	}
}

func itemNameWithAmount(itemName string, amount int32) string {
	if amount == 1 {
		return itemName
	}
	return fmt.Sprintf("%d %s", amount, itemName)
}
