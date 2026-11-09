package game

import (
	"fmt"
	"mud/world"
)

var MENU_WORLD_ENTRIES_ITEM = map[string]MenuEntry {
	"take": {
		usage: "take <item> [from <container>]",
		description: "Pick up an item from the surrounding area, or type 'take <item> from <container>' to take an item from a chest or corpse.",
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
}
