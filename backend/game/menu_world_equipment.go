package game

import (
	"fmt"
	"mud/world"
)

var MENU_WORLD_TOPICS_EQUIP = map[string]MenuHelpTopic {
	"equipment": {
		categories: []string { "equipment" },
		lines: []string {
			"=== Equipment ===",
			"Something.",
			"And something else.",
		},
	},
}

var MENU_WORLD_ENTRIES_EQUIP = map[string]MenuEntry {
	"equip": {
		usage: "equip <item> [in <slot>]",
		description: "Equip the specified item [in the specified slot].",
		categories: []string { "equipment" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			// Check if they are in combat
			/*
			if player.character.IsInCombat(gamestate.world) {
				player.send("You cannot equip items while you are in combat.")
				return true
			}
			*/

			itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "in")

			// Determine the item
			itemIndex, result := fuzzyFindInventoryItem(&player.character.Inventory, itemWords)
			if result != FUZZY_RESULT_OK {
				player.send("Cannot find this item in your inventory.")
				return true
			}

			item := player.character.Inventory.Items[itemIndex]

			// Determine the equipment slot
			var slot world.EquipmentSlot
			var err error
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
				if item.IsOneHandedWeapon() || item.IsShield() {
					player.send(fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'main hand' hand or 'off hand'.", item.NameWithCondition()))
					return false
				}

				var slotFound bool
				slot, slotFound = item.EquipmentSlotForItem()
				if !slotFound {
					player.send(fmt.Sprintf("%s cannot be equipped.", item.NameWithCondition()))
					return true
				}
			}

			// Try to equip item
			equippedItem, messages, err := player.character.EquipFromInventory(gamestate.world, itemIndex, slot)
			if err != nil {
				player.send(err.Error())
				return true
			}
			player.send(fmt.Sprintf("You equipped <g>%s</g>.", equippedItem.NameWithCondition()))
			for _, message := range messages {
				player.send(message)
			}

			return true
		},
	},

	"remove": {
		usage: "remove [<item>] [from <slot>]",
		description: "Remove an equipped item by name or by slot.",
		categories: []string { "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// The way this works is that the user can specify a slot OR an item,
			// but there's no reason for them to specify both so I'm not going to
			// bother writing the code for it

			// Check if they are in combat
			/*
			if player.character.IsInCombat(gamestate.world) {
				player.send("You cannot remove items while you are in combat.")
				return true
			}
			*/

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
				slot, err = fuzzyFindEquipmentSlotByItem(&player.character.Equipment, itemWords)
				if err != nil {
					player.send(err.Error())
					return true
				}
			}

			// Unequip the item
			item, err := player.character.UnequipToInventory(slot)
			if err != nil {
				player.send(err.Error())
				return true
			}
			player.send(fmt.Sprintf("You unequipped <g>%s</g>.", item.NameWithCondition()))

			return true
		},
	},

}
