package game

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
			if userSpecifiedContainer {
				// TODO: fuzzy find container
			}

			// Inventory transfer
		},
	},
}
