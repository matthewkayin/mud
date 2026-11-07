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
}
