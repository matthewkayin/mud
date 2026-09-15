package game

import (
	"fmt"
	"strings"
	"slices"
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
			room := &gameState.world.Rooms[playerMob.data.Room]

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
			room := &gameState.world.Rooms[playerMob.data.Room]

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
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]
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
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

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
			playerMob.data.Room = uint(newRoomIndex)

			*player.inbox <- fmt.Sprintf("You moved into %s.", newRoom.Name)
			describeRoomToPlayer(gameState, player, newRoom)
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
			inventorySize := len(playerMob.data.Inventory.Items)

			if inventorySize == 0 {
				*player.inbox <- "There is nothing in your inventory."
				return true
			}

			itemNames := make([]string, 0, inventorySize)
			for _, item := range playerMob.data.Inventory.Items {
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
			itemIndex, findResult := playerMob.data.Inventory.FuzzyFindItem(args)
			if findResult == INVENTORY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- "That item is not in your inventory."
				return true
			}
			if findResult == INVENTORY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- "Could not drop item. Item name is ambiguous."
				return true
			}

			// Remove item from inventory
			droppedItem := playerMob.data.Inventory.RemoveItem(itemIndex)

			// Add item to room
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]
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
			if len(args) < 1 {
				return false
			}

			// Get pointer to room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.data.Room]

			// Find item in room
			itemIndex, findResult := playerRoom.Inventory.FuzzyFindItem(args)
			if findResult == INVENTORY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- "That item is not in this room."
				return true
			}
			if findResult == INVENTORY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- "Could not take item. Item name is ambiguous."
				return true
			}

			// Move item from room to player
			grabbedItem := playerRoom.Inventory.RemoveItem(itemIndex)
			playerMob.data.Inventory.AddItem(grabbedItem)

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
		usage: "equip <item> [in <slot>]",
		description: "Equip the specified item [in the specified slot].",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			// Check for slot
			inIndex := slices.Index(args, "in")
			slotStr := ""
			if inIndex != -1 && inIndex == len(args) - 1 {
				*player.inbox <- "When specifying 'in' you must also specify a slot."
				return false
			}
			if inIndex != -1 {
				slotStr = args[inIndex + 1]
				args = args[:inIndex]
			}

			// Determine the item
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			itemIndex, findResult := playerMob.data.Inventory.FuzzyFindItem(args)
			if findResult == INVENTORY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- fmt.Sprintf("You have no item called '%s'.", strings.Join(args, " "))
				return true
			}
			if findResult == INVENTORY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- "Could not equip item. The item name is ambigous."
			}

			itemPtr := &playerMob.data.Inventory.Items[itemIndex]
			itemData := ITEM_DATA[itemPtr.Id]

			// Check stat requirements
			itemStatRequires := ItemGetStatRequirements(itemPtr)
			if !playerMob.data.Stats.Meets(itemStatRequires) {
				*player.inbox <- fmt.Sprintf("You do not meet the stat requirements to equip %s.", itemData.name)
				return true
			}

			// Determine slot
			var slot EquipmentSlot
			if slotStr != "" {
				var success bool
				slot, success = EquipmentSlotFromCommandString(slotStr)
				if !success {
					*player.inbox <- fmt.Sprintf("'%s' is not a valid equipment slot.", slotStr)
					*player.inbox <- "Valid equipment slots are: 'mainhand', 'offhand', 'outfit', and 'accessory'."
					return true
				}
			} else {
				if itemData.ItemIsOneHanded() {
					*player.inbox <- fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'mainhand' hand or 'offhand'.", itemData.name)
					return false
				}

				var success bool
				slot, success = EquipmentSlotForItemType(itemData.itemType)
				if !success {
					*player.inbox <- fmt.Sprintf("%s cannot be equipped.", itemData.name)
					return true
				}
			}

			// Try to equip item
			unequippedItems, success := playerMob.data.EquippedItems.Equip(slot, *itemPtr)
			if !success {
				*player.inbox <- fmt.Sprintf("%s cannot be equipped to slot %s.", itemData.name, EquipmentSlotToString(slot))
				return true
			}
			*player.inbox <- fmt.Sprintf("You equipped %s.", itemData.name)

			// Remove item from player inventory
			playerMob.data.Inventory.RemoveItem(itemIndex)

			// Add unequipped items to inventory
			for _, item := range unequippedItems {
				player.onItemUnequipped(gameState, item)
				*player.inbox <- fmt.Sprintf("%s was unequipped and added to your inventory", itemData.name)
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

			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Check for "from"
			fromIndex := slices.Index(args, "from")
			var slot EquipmentSlot = EQUIPMENT_SLOT_COUNT
			slotStr := ""
			if fromIndex != -1 && fromIndex == len(args) - 1 {
				*player.inbox <- "When specifying 'from' you must also specify a slot."
				return false
			}

			// Remove by slot
			if fromIndex != -1 {
				slotStr = args[fromIndex + 1]

				slot, slotFound := EquipmentSlotFromCommandString(slotStr)
				if !slotFound {
					*player.inbox <- fmt.Sprintf("'%s' is not a valid equipment slot.", slotStr)
					*player.inbox <- "Valid equipment slots are: 'mainhand', 'offhand', 'outfit', and 'accessory'."
					return true
				}

				itemPtr := playerMob.data.EquippedItems.Get(slot)
				if itemPtr == nil {
					*player.inbox <- fmt.Sprintf("You have nothing equipped in %s.", EquipmentSlotToString(slot))
					return true
				}
			}

			// Remove by item name
			if slot == EQUIPMENT_SLOT_COUNT {
				bestSlotScore := 0
				for slotIndex := range EQUIPMENT_SLOT_COUNT {
					loopSlot := EquipmentSlot(slotIndex)
					itemPtr := playerMob.data.EquippedItems.Get(loopSlot)
					if itemPtr == nil {
						continue
					}

					score := ItemFuzzyFindScore(*itemPtr, args)
					if score == 0 {
						continue
					}

					if bestSlotScore != 0 && score == bestSlotScore {
						*player.inbox <- "Cannot remove item. The item name provided is too ambiguous."
						return true
					}

					if score > bestSlotScore {
						bestSlotScore = score
						slot = loopSlot
					}
				}

				if bestSlotScore == 0 {
					*player.inbox <- fmt.Sprintf("You have no equipped items matching the name '%s'.", strings.Join(args, " "))
					return true
				}
			}

			if slot == EQUIPMENT_SLOT_COUNT {
				panic("The equipment slot should totally be populated by now")
			}

			// Unequip the item
			item, _ := playerMob.data.EquippedItems.Unequip(slot)
			player.onItemUnequipped(gameState, item)
			*player.inbox <- fmt.Sprintf("You unequipped %s.", ITEM_DATA[item.Id].name)

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
	entries["spellbook"] = MenuEntry {
		usage: "spellbook",
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
			if len(args) != 1 {
				return false
			}

			// Check if the spell is already prepared
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			for _, spell := range playerMob.data.Spells {
				spellData := SPELL_DATA[spell]
				if strings.EqualFold(spellData.name, args[0]) {
					*player.inbox <- fmt.Sprintf("You have already prepared %s.", spellData.name)
					return true
				}
			}

			// Check if spell is in spells known
			for _, spell := range player.character.SpellsKnown {
				spellData := SPELL_DATA[spell]
				if strings.EqualFold(spellData.name, args[0]) {
					playerPrepareSpell(gameState, player, spell)
					return true
				}
			}

			// Check if spell is in spells equipped
			for spell, _ := range player.character.SpellsEquipped {
				spellData := SPELL_DATA[spell]
				if strings.EqualFold(spellData.name, args[0]) {
					playerPrepareSpell(gameState, player, spell)
					return true
				}
			}

			*player.inbox <- fmt.Sprintf("You don't know any spells called '%s'.", args[0])
			return true
		},
	}

	// Forget
	entries["forget"] = MenuEntry {
		usage: "forget <spell>",
		description: "Removes a spell from your prepared spells list. If you have mastered the spell, it will remain in your known spells list.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// Find a spell that matches their input and remove it
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			for _, spell := range playerMob.data.Spells {
				spellData := SPELL_DATA[spell]
				if strings.EqualFold(spellData.name, args[0]) {
					playerMob.data.RemoveSpell(spell)
					*player.inbox <- fmt.Sprintf("You forgot %s.", spellData.name)
					return true
				}
			}

			*player.inbox <- fmt.Sprintf("You haven't prepared any spells called '%s'.", args[0])
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
			for _, spell := range playerMob.data.Spells {
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
			room := &gameState.world.Rooms[playerMob.data.Room]
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
			otherPlayerNames = append(otherPlayerNames, mob.data.Name)
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
	playerRoom := gameState.world.Rooms[playerMob.data.Room]

	for _, occupantHandle := range playerRoom.occupants {
		occupant := gameState.world.Mobs.Get(occupantHandle)
		if strings.EqualFold(occupant.data.Name, targetName) {
			return occupantHandle, true
		}
	}

	return MobHandle{}, false
}

func playerPrepareSpell(gameState *GameState, player *Player, spell Spell) {
	playerMob := gameState.world.Mobs.Get(player.mobHandle)
	playerMob.data.Spells = append(playerMob.data.Spells, spell)
	*player.inbox <- fmt.Sprintf("You prepared %s.", SPELL_DATA[spell].name)
}
