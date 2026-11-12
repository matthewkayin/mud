package game

import (
	"fmt"
	"mud/util"
	"mud/world"
	"strings"
)

var MENU_WORLD_TOPICS_ITEM = map[string]MenuHelpTopic {
	"transfers": {
		categories: []string { "items" },
		lines: []string {
			"=== Item Transfers ===",
			"The commands 'take', 'put', 'drop', and 'give' all allow you to transfer items from one place to another. These commands all use the same syntax for referring to items. That syntax is described here.",
			"1. Item names may contain multiple words. To refer to such items, you can type the name out in full, spaces included. For example 'drop potion of health'.",
			"2. If a command requires two pieces of information, then a delimeter keyword is used. For example, to take a potion from a chest you can say 'take potion of health from chest'. In this case, the 'from' keyword is a delimeter that divides up your input between 'item words' and 'chest words'.",
			"3. Item names are fuzzy matched, so you usually don't have to type the name out in full. For example, 'take potion of health' can be replaced with 'take potion' or 'take health potion'. Notice that the order of words can be re-arranged.",
			"4. If your command is ambiguous (for example, if you typed 'take potion' and there is both a health potion and a mana potion in the room), then you will be told as much and will have to specify the item in more detail.",
			"5. If there are multiple items of the same name, you can choose which one to select by specifying a number after the item name. For example, 'take potion 2 from chest'.",
			"6. If you want to transfer multiple items at once, you can specify a number before the item name. For example, 'put 2 potion into chest'. You can also use the word 'all' as your number. For example, 'take all gold from chest' or 'drop all potion'. Note that this only works with stacking items. Non-stacking items, such as swords, must be transfered one at a time.",
		},
	},
}

var MENU_WORLD_ENTRIES_ITEM = map[string]MenuEntry {
	"take": {
		usage: "take <all|item> [from <container>]",
		description: "Pick up an item from the surrounding area. Type 'take <item> from <container>' to take an item from a chest or corpse. Type 'take all' or 'take all from <container>' to take all items from the room or a container.",
		categories: []string { "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// Get item and container words
			itemWords, containerWords, userSpecifiedContainer := splitArgsBy(args, "from")
			if userSpecifiedContainer && len(containerWords) == 0 {
				return false
			}

			// Determine target inventory
			room := &gamestate.world.Rooms[player.character.Room]
			targetInventory := &room.Inventory
			containerName := "the room"
			if userSpecifiedContainer {
				// TODO: fuzzy find container
			}

			// Special case, if item words is 'all', then loot the container
			if len(itemWords) == 1 && strings.EqualFold(itemWords[0], "all") {
				// Check that target inventory has items
				if len(targetInventory.Items) == 0 {
					player.send(fmt.Sprintf("%s is empty.", containerName))
					return true
				}

				// Check player inventory capactiy
				if !player.character.Inventory.HasSpaceFor(targetInventory.Size(), world.INVENTORY_CAPACITY_PLAYER) {
					player.send("You don't have enough space in your inventory for all that.")
					return true
				}

				// Build item names
				itemNames := make([]string, 0, len(targetInventory.Items))
				for len(targetInventory.Items) > 0 {
					item := targetInventory.RemoveStack(gamestate.world, len(targetInventory.Items) - 1)
					player.character.Inventory.AddItem(item)
					itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", item.NameWithAmount()))
				}

				player.send(fmt.Sprintf("You took %s from %s.", util.CombineNames(itemNames), containerName))
				player.isInventoryDirty = true
				return true
			}

			// Inventory transfer
			result, err := gamestate.inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: targetInventory,
					name: containerName,
					capacity: world.INVENTORY_CAPACITY_UNLIMITED,
				},
				to: InventoryTransferLocation {
					inventory: &player.character.Inventory,
					name: "your inventory",
					capacity: world.INVENTORY_CAPACITY_PLAYER,
				},
				verb: "take",
				itemWords: itemWords,
			})
			if err != nil {
				player.send(err.Error())
				return true
			}
			if result.notice != "" {
				player.send(result.notice)
			}

			player.send(fmt.Sprintf("You took <g>%s</g> from %s.", result.item.NameWithAmount(), containerName))
			player.isInventoryDirty = true
			return true
		},
	},

	"put": {
		usage: "put <item> into <container>",
		description: "Puts an item from your inventory into the specified container.",
		categories: []string { "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// Get item and container words
			itemWords, containerWords, userSpecifiedContainer := splitArgsBy(args, "into")
			if !userSpecifiedContainer || len(itemWords) == 0 || len(containerWords) == 0 {
				return false
			}

			// Find target inventory
			// TODO: replace with fuzzyFindChestInventory
			room := &gamestate.world.Rooms[player.character.Room]
			containerName := "the room"
			targetInventory := &room.Inventory

			// Transfer
			result, err := gamestate.inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &player.character.Inventory,
					name: "your inventory",
					capacity: world.INVENTORY_CAPACITY_PLAYER,
				},
				to: InventoryTransferLocation {
					inventory: targetInventory,
					name: containerName,
					capacity: world.INVENTORY_CAPACITY_UNLIMITED,
				},
				verb: "put",
				itemWords: itemWords,
			})
			if err != nil {
				player.send(err.Error())
				return true
			}
			if result.notice != "" {
				player.send(result.notice)
			}

			player.send(fmt.Sprintf("You put <g>%s</g> into %s.", result.item.NameWithAmount(), containerName))
			player.isInventoryDirty = true
			return true
		},
	},

	"give": {
		usage: "give <item> to <target>",
		description: "Gives an item to the target",
		categories: []string { "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// Split args
			itemWords, targetWords, userSpecifiedTo := splitArgsBy(args, "to")
			if !userSpecifiedTo || len(itemWords) == 0 || len(targetWords) == 0 {
				return false
			}

			// Find target
			targetHandle, err := fuzzyFindMob(gamestate, player, targetWords)
			if err != nil {
				player.send(err.Error())
				return true
			}

			// Transfer
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			result, err := gamestate.inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &player.character.Inventory,
					name: "your inventory",
					capacity: world.INVENTORY_CAPACITY_PLAYER,
				},
				to: InventoryTransferLocation {
					inventory: &targetMob.Inventory,
					name: targetMob.Name,
					capacity: targetMob.InventoryCapacity(),
				},
				verb: "give",
				itemWords: itemWords,
			})
			if err != nil {
				player.send(err.Error())
				return true
			}
			if result.notice != "" {
				player.send(result.notice)
			}

			gamestate.world.MessageRoom(player.character.Room, fmt.Sprintf("<y>%s</y> gave <g>%s</g> to <y>%s</y>",
				player.character.Name, result.item.NameWithAmount(), targetMob.Name))
			player.isInventoryDirty = true
			if targetMob.IsPlayer() {
				targetPlayerIndex := gamestate.playerKeyToIndexMap[targetMob.Character.PlayerKey]
				gamestate.players[targetPlayerIndex].isInventoryDirty = true
			}
			return true
		},
	},

	"drop": {
		usage: "drop <item>",
		description: "Drop an item from your inventory onto the floor",
		categories: []string { "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			room := &gamestate.world.Rooms[player.character.Room]

			// Transfer
			result, err := gamestate.inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &player.character.Inventory,
					name: "your inventory",
					capacity: world.INVENTORY_CAPACITY_PLAYER,
				},
				to: InventoryTransferLocation {
					inventory: &room.Inventory,
					name: "the room",
					capacity: world.INVENTORY_CAPACITY_UNLIMITED,
				},
				verb: "drop",
				itemWords: args,
			})
			if err != nil {
				player.send(err.Error())
				return true
			}
			if result.notice != "" {
				player.send(result.notice)
			}

			player.send(fmt.Sprintf("You dropped <g>%s</g>.", result.item.NameWithAmount()))
			player.isInventoryDirty = true
			return true
		},
	},
}
