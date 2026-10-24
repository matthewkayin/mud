package game

import (
	"fmt"
	"mud/world"
)

var MENU_WORLD_ENTRIES_EQUIP = map[string]MenuEntry {
	"equipment": {
		usage: "equipment",
		description: "Show your current equipment",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

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

			// Check if they are in combat
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.IsInCombat(gamestate.world) {
				player.send("You cannot equip items while you are in combat.")
				return true
			}

			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "in")

			// Determine the item
			itemIndex, err := fuzzyFindInventoryItem(&playerMob.Data.Inventory, "your inventory", itemWords)
			if err != nil {
				player.send(err.Error())
				return true
			}
			item := &playerMob.Data.Inventory.Items[itemIndex]
			itemData := world.ITEM_DATA[item.Id]

			// Determine the equipment slot
			var slot world.EquipmentSlot
			if userSpecifiedSlot {
				if len(slotWords) == 0 {
					player.send("When specifying 'in' you must also specify a slot.")
					return false
				}

				slot, err = fuzzyFindEquipmentSlot(slotWords)
				if err != nil {
					player.send(err.Error())
					return true
				}
			} else {
				if itemData.ItemIsOneHanded() {
					player.send(fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'main hand' hand or 'off hand'.", item.GetNameWithCondition()))
					return false
				}

				var slotFound bool
				slot, slotFound = world.EquipmentSlotForItemKind(itemData.Kind)
				if !slotFound {
					player.send(fmt.Sprintf("%s cannot be equipped.", item.GetNameWithCondition()))
					return true
				}
			}

			// Try to equip item
			equippedItem, messages, err := playerMob.EquipFromInventory(itemIndex, slot)
			if err != nil {
				player.send(err.Error())
				return true
			}
			player.send(fmt.Sprintf("You equipped <g>%s</g>.", equippedItem.GetNameWithCondition()))
			for _, message := range messages {
				player.send(message)
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

			// Check if they are in combat
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.IsInCombat(gamestate.world) {
				player.send("You cannot remove items while you are in combat.")
				return true
			}

			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "from")

			var slot world.EquipmentSlot
			if userSpecifiedSlot {
				if len(itemWords) != 0 {
					player.send("When specifying 'from', you should not specify an item.")
					return false
				}
				if len(slotWords) == 0 {
					player.send("When specifying 'from', you must also specify a slot.")
					return false
				}

				// Find slot
				var err error
				slot, err = fuzzyFindEquipmentSlot(slotWords)
				if err != nil {
					player.send(err.Error())
					return true
				}
			} else {
				// Find slot
				var err error
				slot, err = fuzzyFindEquipmentSlotByItem(&playerMob.Data.Equipment, itemWords)
				if err != nil {
					player.send(err.Error())
					return true
				}
			}

			// Unequip the item
			item, messages, err := playerMob.UnequipToInventory(slot)
			if err != nil {
				player.send(err.Error())
				return true
			}
			player.send(fmt.Sprintf("You unequipped <g>%s</g>.", item.GetNameWithCondition()))
			for _, message := range messages {
				player.send(message)
			}

			return true
		},
	},

}
