package game

import (
	"fmt"
	"strings"
)

func MenuWorld() Menu {
	entries := make(map[string]MenuEntry)

	// Logout
	entries["logout"] = MenuEntry {
		usage: "logout",
		description: "Logout of the world.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// TODO: remove player from current room
			// TODO: broadcast world message to everyone who is logged in? or just to the current room?
			player.enterMenu(gameState, &gameState.menuLogin)
			return true
		},
	}

	// Look
	entries["look"] = MenuEntry {
		usage: "look",
		description: "Describe the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			room := &gameState.world.rooms[player.character.currentRoom]

			*(player.inbox) <- room.description

			// Send the list of players in the room
			if len(room.occupants) > 1 {
				otherPlayerCount := len(room.occupants) - 1

				otherPlayerNames := make([]string, 0, otherPlayerCount)
				for _, otherPlayerId := range room.occupants {
					// Don't tell the player about themselves being in the room
					if otherPlayerId == player.id {
						continue
					}

					// Get a handle to the other player
					otherPlayerIndex := gameState.playerIdToIndexMap[otherPlayerId]
					otherPlayer := &gameState.players[otherPlayerIndex]

					// Add their name to the list
					otherPlayerNames = append(otherPlayerNames, otherPlayer.character.name)
				}

				otherPlayersStr := menuWorldCombineNames(otherPlayerNames)
				*(player.inbox) <- fmt.Sprintf("Players in this room: %s", otherPlayersStr)
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
				*(player.inbox) <- "You must include a message that you want to say."
				return false
			}

			gameState.broadcast(fmt.Sprintf("%s: '%s'", player.character.name, strings.Join(args, " ")))
			return true
		},
	}

	return Menu {
		entries: entries,
		getDescription: func (gameState *GameState, player *Player) string {
			room := &gameState.world.rooms[player.character.currentRoom]
			return fmt.Sprintf("You are in %s.", room.name)
		},
	}
}

func menuWorldCombineNames(names []string) string {
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
