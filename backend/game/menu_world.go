package game

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"mud/bitset"
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
			direction, directionFound := DirectionFromString(args[0])
			if !directionFound {
				*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
				return false
			}

			newRoomIndex := playerRoom.Exits[direction]

			// Check to make sure there is an exit
			if newRoomIndex == ROOM_NONE {
				*player.inbox <- "There is not exit in that direction."
				return true
			}

			// Check to make sure the door is not locked
			if playerRoom.ExitIsLocked[direction] {
				*player.inbox <- fmt.Sprintf("The %s exit is locked!", DirectionToString(direction))
				return true
			}

			// Get a pointer to the new room
			newRoom := &gameState.world.Rooms[newRoomIndex]

			// Move the player
			playerRoom.RemoveOccupant(player.mobHandle)
			newRoom.AddOccupant(player.mobHandle)
			playerMob.data.Room = newRoomIndex
			bitset.Set(player.character.RoomsDiscovered, newRoomIndex, true)

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
			if len(args) < 1 {
				return false
			}

			targetHandle, targetFound := fuzzyFindTarget(gameState, player, args)
			if !targetFound {
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

			*player.inbox <- "Item  | Type | Description "
			for index := range len(playerMob.data.Inventory.Items) {
				item := &playerMob.data.Inventory.Items[index]
				itemData := ITEM_DATA[item.Id]
				itemStatRequirements := ItemGetStatRequirements(item)

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

				*player.inbox <- fmt.Sprintf("%s | %s | %s", itemData.name, typeStr, itemData.description)
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
			itemIndex := fuzzyFindInventoryItemIndex(&playerMob.data.Inventory, args)

			// Handle edge cases
			if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- fmt.Sprintf("No item called '%s' is in your inventory.",
					strings.Join(args, " "))
				return true
			}
			if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- fmt.Sprintf("The item name '%s' is ambiguous.",
					strings.Join(args, " "))
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
			itemIndex := fuzzyFindInventoryItemIndex(&playerRoom.Inventory, args)
			if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- fmt.Sprintf("No item called '%s' is in this room.",
					strings.Join(args, " "))
				return true
			}
			if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- fmt.Sprintf("The item name '%s' is ambiguous.",
					strings.Join(args, " "))
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

			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "in")
			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Determine the item
			itemIndex := fuzzyFindInventoryItemIndex(&playerMob.data.Inventory, itemWords)

			// Handle edge cases
			if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- fmt.Sprintf("No item called '%s' is in your inventory.",
					strings.Join(itemWords, " "))
				return true
			}
			if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- fmt.Sprintf("The item name '%s' is ambiguous.",
					strings.Join(itemWords, " "))
				return true
			}

			// Check stat requirements
			item := &playerMob.data.Inventory.Items[itemIndex]
			itemData := ITEM_DATA[item.Id]
			itemStatRequires := ItemGetStatRequirements(item)
			if !playerMob.data.Stats.Meets(itemStatRequires) {
				*player.inbox <- fmt.Sprintf("You do not meet the stat requirements to equip %s.", itemData.name)
				return true
			}

			// Determine the equipment slot
			var slot EquipmentSlot
			var slotFound bool
			if userSpecifiedSlot {
				if len(slotWords) == 0 {
					*player.inbox <- "When specifying 'in' you must also specify a slot."
					return false
				}

				slot, slotFound = fuzzyFindEquipmentSlot(player, slotWords)
				if !slotFound {
					return true
				}
			} else {
				if itemData.ItemIsOneHanded() {
					*player.inbox <- fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'mainhand' hand or 'offhand'.", itemData.name)
					return false
				}

				slot, slotFound = EquipmentSlotForItemType(itemData.itemType)
				if !slotFound {
					*player.inbox <- fmt.Sprintf("%s cannot be equipped.", itemData.name)
					return true
				}
			}

			// Try to equip item
			unequippedItems, success := playerMob.data.EquippedItems.Equip(slot, *item)
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

			// The way this works is that the user can specify a slot OR an item,
			// but there's no reason for them to specify both so I'm not going to
			// bother writing the code for it

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "from")

			var slot EquipmentSlot
			var slotFound bool
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
				slot, slotFound = fuzzyFindEquipmentSlot(player, slotWords)
				if !slotFound {
					return true
				}
			} else {
				// Find slot
				slot, slotFound = fuzzyFindEquipmentSlotByItem(player, &playerMob.data.EquippedItems, itemWords)
				if !slotFound {
					return true
				}
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
			// Check if the spell is already prepared
			spell, isPrepared := fuzzyFindPreparedSpell(gameState, player, args)
			if isPrepared {
				spellData := SPELL_DATA[spell]
				*player.inbox <- fmt.Sprintf("You have already prepared %s.", spellData.name)
				return true
			}

			// Search for spell
			spell, spellFound := fuzzyFindKnownOrEquippedSpell(player, args)
			if !spellFound {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
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
			spell, isPrepared := fuzzyFindPreparedSpell(gameState, player, args)
			if !isPrepared {
				*player.inbox <- fmt.Sprintf("You haven't prepared any spells called '%s'.",
					strings.Join(args, " "))
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
			spell, spellFound := fuzzyFindPreparedSpell(gameState, player, spellWords)
			if !spellFound {
				return true
			}

			// Find the target in the room
			targetHandle, targetFound := fuzzyFindTarget(gameState, player, targetWords)
			if !targetFound {
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
			if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
				*player.inbox <- fmt.Sprintf("No item called '%s' is in your inventory.",
					strings.Join(args, " "))
				return true
			}
			if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
				*player.inbox <- fmt.Sprintf("The item name '%s' is ambiguous.",
					strings.Join(args, " "))
				return true
			}

			// Determine target
			var targetHandle MobHandle
			var targetFound bool
			if userSpecifiedTarget {
				targetHandle, targetFound = fuzzyFindTarget(gameState, player, targetWords)
				if !targetFound {
					return true
				}
			} else {
				targetHandle = player.mobHandle
			}

			item := &playerMob.data.Inventory.Items[itemIndex]
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
