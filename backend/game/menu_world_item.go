package game

import (
	"fmt"
	"mud/util"
	"mud/world"
)

var MENU_WORLD_ENTRIES_ITEM = map[string]MenuEntry {
	"inventory": {
		usage: "inventory",
		description: "List the items in your inventory",
		handler: func(gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

			if playerMob.Data.Inventory.Length() == 0 {
				player.send("There is nothing in your inventory.")
				return true
			}

			// Inventory table headers
			inventoryHeaders := []string {
				"Item",
				"Amount",
				fmt.Sprintf("Size (%d / %d)", playerMob.Data.Inventory.Size(), playerMob.InventoryCapacity()),
			}
			inventoryRows := make([]string, 0, 3 * playerMob.Data.Inventory.Length())

			// Inventory table rows
			for index := range playerMob.Data.Inventory.Length() {
				item := &playerMob.Data.Inventory.Items[index]
				itemData := world.ITEM_DATA[item.Id]

				inventoryRows = append(inventoryRows, itemData.Name)
				inventoryRows = append(inventoryRows, fmt.Sprintf("x%d", item.Amount))

				if item.Amount == 1 || itemData.Size == 0 {
					inventoryRows = append(inventoryRows, fmt.Sprintf("%d", itemData.Size))
				} else {
					inventoryRows = append(inventoryRows, fmt.Sprintf("%d (%dx%d)", itemData.Size * item.Amount, itemData.Size, item.Amount))
				}
			}

			printTable(player, "Inventory", inventoryHeaders, inventoryRows)

			player.send("\nType 'inspect <item>' to see more details about an item.")
			return true
		},
	},

	"drop": {
		usage: "drop <item>",
		description: "Drop an item from your inventory",
		handler: func(gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			result, err := inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &playerMob.Data.Inventory,
					name: "your inventory",
					capacity: playerMob.InventoryCapacity(),
				},
				to: InventoryTransferLocation {
					inventory: &playerRoom.Inventory,
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
			player.send(fmt.Sprintf("You dropped <g>%s</g>.", result.item.GetNameWithAmount()))

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
				player.send(err.Error())
				return true
			}

			// Transfer
			result, err := inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &playerMob.Data.Inventory,
					name: "your inventory",
					capacity: playerMob.InventoryCapacity(),
				},
				to: InventoryTransferLocation {
					inventory: targetInventory,
					name: chestName,
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
			player.send(fmt.Sprintf("You put <g>%s</g> into %s.", result.item.GetNameWithAmount(), chestName))

			return true
		},
	},

	"give": {
		usage: "give <item> to <target>",
		description: "Gives an item to the target",
		handler: func(gamestate *GameState, player *Player, args []string) bool {
			// Split args
			itemWords, targetWords, userSpecifiedTo := splitArgsBy(args, "to")
			if !userSpecifiedTo || len(itemWords) == 0 || len(targetWords) == 0 {
				return false
			}

			// Find target
			targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
			if err != nil {
				player.send(err.Error())
				return true
			}

			// Don't allow give to player because players could maliciously fill up another player's inventory
			// This makes me think we should drop the "give" command entirely, but it's used for troll behavior
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.PlayerCharacter != nil {
				player.send(fmt.Sprintf("You cannot give items to %s because they are a player. Request to 'trade' with them instead.", targetMob.Data.Name))
				return true
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

			result, err := inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &playerMob.Data.Inventory,
					name: "your inventory",
					capacity: playerMob.InventoryCapacity(),
				},
				to: InventoryTransferLocation {
					inventory: &targetMob.Data.Inventory,
					name: targetMob.Data.Name,
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
			gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> gave <g>%s</g> to <y>%s</y>.",
				playerMob.Data.Name, result.item.GetNameWithAmount(), targetMob.Data.Name))

			if targetMob.Npc != nil {
				targetMob.Npc.FireEvent(gamestate.world, world.NpcEvent {
					Type: world.NPC_EVENT_TYPE_ITEM_GIVEN,
					Data: world.NpcEventItemGiven {
						PlayerHandle: player.mobHandle,
						AddedToIndex: result.addedToIndex,
						Amount: result.item.Amount,
					},
				})
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
					player.send(err.Error())
					return true
				}
			}

			// Inventory transfer
			result, err := inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: targetInventory,
					name: chestName,
					capacity: world.INVENTORY_CAPACITY_UNLIMITED,
				},
				to: InventoryTransferLocation {
					inventory: &playerMob.Data.Inventory,
					name: "your inventory",
					capacity: playerMob.InventoryCapacity(),
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
			player.send(fmt.Sprintf("You took <g>%s</g> from %s.", result.item.GetNameWithAmount(), chestName))

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
				player.send(err.Error())
				return true
			}

			if len(targetInventory.Items) == 0 {
				player.send(fmt.Sprintf("%s is empty.", chestName))
				return true
			}

			if !playerMob.Data.Inventory.HasSpaceFor(targetInventory.Size(), playerMob.InventoryCapacity()) {
				player.send("You don't have enough space in your inventory for all that.")
				return true
			}

			itemNames := make([]string, 0, len(targetInventory.Items))
			for len(targetInventory.Items) > 0 {
				item := targetInventory.RemoveStack(len(targetInventory.Items) - 1)

				playerMob.Data.Inventory.AddItem(item)
				itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", item.GetNameWithAmount()))
			}
			player.send(fmt.Sprintf("You got %s.", util.CombineNames(itemNames)))

			return true
		},
	},
}
