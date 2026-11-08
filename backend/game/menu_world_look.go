package game

import (
	"fmt"
	"mud/util"
	"strings"
)

var MENU_WORLD_ENTRIES_LOOK = map[string]MenuEntry {
	"look": {
		usage: "look [at <detail|target|item>]",
		description: "Describe the current room, or 'look at' something to get information about it.",
		categories: []string { "exploration" },
		handler: func(gamestate *GameState, player *Player, args []string) bool {
			room := &gamestate.world.Rooms[player.character.Room]

			// If no args, then do a regular look
			if len(args) == 0 {
				player.send(fmt.Sprintf("You are in <y>%s</y>.", room.Name))
				describeRoomToPlayer(gamestate, player, room)
				return true
			}

			// If args, then make sure they specified "at <something>"
			if !strings.EqualFold(args[0], "at") || len(args) == 1 {
				return false
			}

			searchWords := strings.ToLower(strings.Join(args[1:], " "))

			// TODO: check for target mob or item

			// Check for room detail
			description, descriptionFound := room.Details[searchWords]
			if descriptionFound {
				player.send(description)
				return true
			}

			// Fallback to matching on the room description
			sentence, found := util.FindSentenceContainingSubstr(room.Description, searchWords)
			if found {
				player.send(sentence)
				return true
			}

			player.send(fmt.Sprintf("You don't see any '%s' in the room.", searchWords))
			return true
		},
	},

	"search": {
		usage: "search [container]",
		description: "Search the room for items, or type 'search <container>' to search a specific container such as a chest or a corpse.",
		categories: []string { "exploration", "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// TODO: make this do a 'search' role so that some players are better at searching than others

			room := &gamestate.world.Rooms[player.character.Room]
			targetInventory := &room.Inventory

			// TODO: check for target inventory

			// Handle no items case
			if len(targetInventory.Items) == 0 {
				player.send("You see nothing.")
				return true
			}

			// Build a list of item names
			itemNames := make([]string, 0, len(targetInventory.Items))
			for index := range len(targetInventory.Items) {
				itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", targetInventory.Items[index].Name))
			}

			player.send(fmt.Sprintf("You see %s.", util.CombineNames(itemNames)))
			return true
		},
	},
}
