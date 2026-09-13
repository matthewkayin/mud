package game

import (
	"fmt"
	"strings"
	"log"
)

func MenuWorld() Menu {
	entries := make(map[string]MenuEntry)

	// Logout
	entries["logout"] = MenuEntry {
		usage: "logout",
		description: "Logout of the world.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			player.exitWorld(gameState)
			return true
		},
	}

	// Look
	entries["look"] = MenuEntry {
		usage: "look",
		description: "Describe the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]

			describeRoomToPlayer(gameState, player, room)
			return true
		},
	}

	// Exits
	entries["exits"] = MenuEntry {
		usage: "exits",
		description: "Describe the exits of the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]

			printRoomExits(gameState, player, room)
			return true
		},
	}

	// Say
	entries["say"] = MenuEntry {
		usage: "say <message>",
		description: "Send a messsage to the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				*player.inbox <- "You must include a message that you want to say."
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
			playerRoom.broadcast(gameState, fmt.Sprintf("%s said '%s'", player.character.Data.Name, strings.Join(args, " ")))
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
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]

			// Determine the index of the target room
			newRoomIndex := ROOM_NONE
			switch strings.ToLower(args[0]) {
				case "north":
					newRoomIndex = playerRoom.ExitNorth
				case "south":
					newRoomIndex = playerRoom.ExitSouth
				case "east":
					newRoomIndex = playerRoom.ExitEast
				case "west":
					newRoomIndex = playerRoom.ExitWest
				default:
					*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
					return false
			}

			// Check to make sure there is an exit
			if newRoomIndex == ROOM_NONE {
				*player.inbox <- "There is not exit in that direction."
				return true
			}

			// Get a pointer to the new room
			newRoom := &gameState.world.Rooms[newRoomIndex]

			// Move the player
			playerRoom.RemoveOccupant(player.mobHandle)
			newRoom.AddOccupant(player.mobHandle)
			playerMob.Data.Room = uint(newRoomIndex)

			*player.inbox <- fmt.Sprintf("You moved into %s.", newRoom.Name)
			describeRoomToPlayer(gameState, player, newRoom)
			return true
		},
	}

	// Status
	entries["status"] = MenuEntry {
		usage: "status",
		description: "Show your current status.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			player.printStatus(gameState)
			return true
		},
	}

	// Attack
	entries["attack"] = MenuEntry {
		usage: "attack <target>",
		description: "Attack the target.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			targetHandle, targetFound := getTargetHandle(gameState, playerMob, args[0])
			if !targetFound {
				*player.inbox <- fmt.Sprintf("No target named '%s' is in this room.", args[0])
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
			inventorySize := len(playerMob.Data.Inventory.Items)

			if inventorySize == 0 {
				*player.inbox <- "There is nothing in your inventory."
				return true
			}

			itemNames := make([]string, 0, inventorySize)
			for _, item := range playerMob.Data.Inventory.Items {
				itemNames = append(itemNames, ITEM_DATA[item.Id].name)
			}
			*player.inbox <- fmt.Sprintf("You are carrying the following items: %s", combineNames(itemNames))
			return true
		},
	}

	// Drop an item
	entries["drop"] = MenuEntry{
		usage: "drop <item>",
		description: "Drop an item from your inventory",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Find item
			itemIndex, hasItem := playerMob.Data.Inventory.FindItem(args[0])
			if !hasItem {
				*player.inbox <- "That item is not in your inventory."
				return true
			}

			// Remove item from inventory
			droppedItem := playerMob.Data.Inventory.RemoveItem(itemIndex)

			// Add item to room
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
			playerRoom.Inventory.AddItem(droppedItem)

			*player.inbox <- fmt.Sprintf("You dropped %s.", ITEM_DATA[droppedItem.Id].name)
			return true
		},
	}

	// Grab an item
	entries["take"] = MenuEntry {
		usage: "take <item>",
		description: "Pick up an item in your current room",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// Get pointer to room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]

			// Find item in room
			itemIndex, hasItem := playerRoom.Inventory.FindItem(args[0])
			if !hasItem {
				*player.inbox <- "That item is not in this room."
				return true
			}

			// Move item from room to player
			grabbedItem := playerRoom.Inventory.RemoveItem(itemIndex)
			playerMob.Data.Inventory.AddItem(grabbedItem)

			*player.inbox <- fmt.Sprintf("You picked up %s.", ITEM_DATA[grabbedItem.Id].name)
			return true
		},
	}

	// Show equipment
	entries["equipment"] = MenuEntry {
		usage: "equipment",
		description: "Show your current equipment",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			log.Printf("Handling equipment")
			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Determine if we should skip the offhand item slot
			mainHandItem := playerMob.Data.EquippedItems.Get(EQUIPMENT_SLOT_MAIN_HAND)
			shouldSkipOffhand := mainHandItem != nil && ITEM_DATA[mainHandItem.Id].itemType == EQUIPMENT_SLOT_MAIN_HAND

			*player.inbox <- "Your equipment is:"

			for index := range EQUIPMENT_SLOT_COUNT {
				slot := EquipmentSlot(index)
				item := playerMob.Data.EquippedItems.Get(slot)
				var itemData *ItemData = nil

				// If two handed equipped, skip off hand
				if slot == EQUIPMENT_SLOT_OFF_HAND && shouldSkipOffhand {
					continue
				}

				// Determine item name
				var itemName string
				if item != nil {
					itemData = ITEM_DATA[item.Id]
					itemName = itemData.name
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
		usage: "equip <item> [slot]",
		description: "Equip the specified item",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 || len(args) > 2 {
				return false
			}

			// Determine the item
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			itemIndex, itemFound := playerMob.Data.Inventory.FindItem(args[0])
			if !itemFound {
				*player.inbox <- fmt.Sprintf("You have no item called '%s'.", args[0])
				return true
			}
			itemPtr := &playerMob.Data.Inventory.Items[itemIndex]
			itemData := ITEM_DATA[itemPtr.Id]

			// Determine slot
			var slot EquipmentSlot
			if len(args) == 2 {
				var success bool
				slot, success = EquipmentSlotFromCommandString(args[1])
				if !success {
					*player.inbox <- fmt.Sprintf("'%s' is not a valid equipment slot.", args[1])
					*player.inbox <- "Valid equipment slots are: 'mainhand', 'offhand', 'helm', 'armor', 'boots', and 'accessory'."
					return true
				}
			} else {
				if itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED {
					*player.inbox <- fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'mainhand' or 'offhand'.", itemData.name)
					return true
				}

				var success bool
				slot, success = EquipmentSlotForItemType(itemData.itemType)
				if !success {
					*player.inbox <- fmt.Sprintf("%s cannot be equipped.", itemData.name)
					return true
				}
			}

			// Try to equip item
			unequippedItems, success := playerMob.Data.EquippedItems.Equip(slot, *itemPtr)
			if !success {
				*player.inbox <- fmt.Sprintf("%s cannot be equipped to slot %s.", itemData.name, EquipmentSlotToString(slot))
				return true
			}

			*player.inbox <- fmt.Sprintf("You equipped %s.", itemData.name)

			// Remove item from player inventory
			playerMob.Data.Inventory.RemoveItem(itemIndex)

			// Add unequipped items to inventory
			for _, item := range unequippedItems {
				playerMob.Data.Inventory.AddItem(item)
				unequippedItemData := ITEM_DATA[item.Id]
				*player.inbox <- fmt.Sprintf("%s was unequipped and added to your inventory.", unequippedItemData.name)
			}

			return true
		},
	}

	// Unequip
	entries["remove"] = MenuEntry {
		usage: "remove <item>",
		description: "Remove an equipped item",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Find the slot
			var slot EquipmentSlot
			foundItem := false
			for slotIndex := range EQUIPMENT_SLOT_COUNT {
				slot = EquipmentSlot(slotIndex)
				itemPtr := playerMob.Data.EquippedItems.Get(slot)
				if itemPtr == nil {
					continue
				}

				itemData := ITEM_DATA[itemPtr.Id]
				if strings.EqualFold(itemData.name, args[0]) {
					foundItem = true
					break
				}
			}

			// Return early if no item found
			if !foundItem {
				*player.inbox <- fmt.Sprintf("You have no equipped items called '%s'.", args[0])
				return true
			}

			// Unequip the item
			item, _ := playerMob.Data.EquippedItems.Unequip(slot)
			playerMob.Data.Inventory.AddItem(item)
			*player.inbox <- fmt.Sprintf("You unequipped %s.", ITEM_DATA[item.Id].name)

			return true
		},
	}

	// Spell list
	entries["spells"] = MenuEntry {
		usage: "spells",
		description: "Show a list of the spells that you know",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			if len(playerMob.Data.Spells) == 0 {
				*player.inbox <- "You don't know any spells."
				return true
			}

			for _, spell := range playerMob.Data.Spells {
				spellData := SPELL_DATA[spell]
				*player.inbox <- fmt.Sprintf("%s (Mana Cost: %d) - %s", spellData.name, spellData.manaCost, spellData.description)
			}

			return true
		},
	}

	// Cast
	entries["cast"] = MenuEntry {
		usage: "cast <spell> <target>",
		description: "Cast a spell",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 2 {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Find the spell in their spell list
			var spellToCast Spell
			spellFound := false
			for _, spell := range playerMob.Data.Spells {
				spellData := SPELL_DATA[spell]
				if strings.EqualFold(spellData.name, args[0]) {
					spellToCast = spell
					spellFound = true
				}
			}
			if !spellFound {
				*player.inbox <- fmt.Sprintf("You don't know of any spells called '%s'.", args[0])
				return true
			}

			// Find the target in the room
			targetHandle, targetFound := getTargetHandle(gameState, playerMob, args[1])
			if !targetFound {
				*player.inbox <- fmt.Sprintf("No target named '%s' is in this room.", args[1])
				return true
			}

			// Queue up a cast action
			player.nextAction = Action {
				actionType: ACTION_TYPE_CAST,
				data: ActionCast {
					spell: spellToCast,
					target: targetHandle,
				},
			}

			return true
		},
	}

	return Menu {
		entries: entries,
		getDescription: func (gameState *GameState, player *Player) string {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]
			return fmt.Sprintf("You are in %s.", room.Name)
		},
	}
}

func combineNames(names []string) string {
	switch len(names) {
		case 0:
			return ""
		case 1:
			return names[0]
		case 2:
			return names[0] + " and " + names[1]
		default:
			return strings.Join(names[:len(names) - 1], ", ") + ", and " + names[len(names) - 1]
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
			if mobHandle.Equals(player.mobHandle) {
				continue
			}

			// Get a pointer to the mob
			mob := gameState.world.Mobs.Get(mobHandle)
			// Add their name to the list
			otherPlayerNames = append(otherPlayerNames, mob.Data.Name)
		}

		otherPlayersStr := combineNames(otherPlayerNames)
		isString := "are"
		if otherPlayerCount == 1 {
			isString = "is"
		}
		*player.inbox <- fmt.Sprintf("%s %s here.", otherPlayersStr, isString)
	}

	if len(room.Inventory.Items) > 0 {
		itemNames := make([]string, 0, len(room.Inventory.Items))
		for _, item := range room.Inventory.Items {
			itemNames = append(itemNames, ITEM_DATA[item.Id].name)
		}
		isString := "items are"
		if len(room.Inventory.Items) == 1 {
			isString = "item is"
		}
		*player.inbox <- fmt.Sprintf("The following %s in this room: %s.", isString, combineNames(itemNames))
	}
}

func printRoomExits(gameState *GameState, player *Player, room *Room) {
	exitFound := false

	if room.ExitNorth != ROOM_NONE {
		exitRoom := &gameState.world.Rooms[room.ExitNorth]
		*player.inbox <- fmt.Sprintf("To the north is %s", exitRoom.Name)
		exitFound = true
	}

	if room.ExitSouth != ROOM_NONE {
		exitRoom := &gameState.world.Rooms[room.ExitSouth]
		*player.inbox <- fmt.Sprintf("To the south is %s", exitRoom.Name)
		exitFound = true
	}

	if room.ExitEast != ROOM_NONE {
		exitRoom := &gameState.world.Rooms[room.ExitEast]
		*player.inbox <- fmt.Sprintf("To the east is %s", exitRoom.Name)
		exitFound = true
	}

	if room.ExitWest != ROOM_NONE {
		exitRoom := &gameState.world.Rooms[room.ExitWest]
		*player.inbox <- fmt.Sprintf("To the west is %s", exitRoom.Name)
		exitFound = true
	}

	if !exitFound {
		*player.inbox <- "This room has no exits!"
	}
}

func getTargetHandle(gameState *GameState, playerMob *Mob, targetName string) (MobHandle, bool) {
	playerRoom := gameState.world.Rooms[playerMob.Data.Room]

	for _, occupantHandle := range playerRoom.occupants {
		occupant := gameState.world.Mobs.Get(occupantHandle)
		if strings.EqualFold(occupant.Data.Name, targetName) {
			return occupantHandle, true
		}
	}

	return MobHandle{}, false
}
