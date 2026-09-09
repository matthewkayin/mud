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
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			return true
		},
	}

	// Look
	entries["look"] = MenuEntry {
		usage: "look",
		description: "Describe the current room.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			room := &gameState.world.rooms[player.character.currentRoom]

			*(player.inbox) <- room.description

			if len(room.playersInRoom) > 1 {
				playersInRoom := "The following players are in this room: "
				for index, inRoomPlayerId := range room.playersInRoom {
					if index == len(room.playersInRoom) - 1 {
						playersInRoom += "and "
					}
					playersInRoom += gameState.players[gameState.playerIdToIndexMap[inRoomPlayerId]].character.name
					if index < len(room.playersInRoom) - 1 {
						playersInRoom += ", "
					}
					if index == len(room.playersInRoom) - 1 {
						playersInRoom += "."
					}
				}
				*(player.inbox) <- playersInRoom
			}

			return true
		},
	}

	// Say
	entries["say"] = MenuEntry {
		usage: "say <message>",
		description: "Send a messsage to the current room.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			if len(args) < 1 {
				*(player.inbox) <- "You must include a message that you want to say."
				return false
			}

			gameState.broadcast(fmt.Sprintf("%s: \"%s\"", player.character.name, strings.Join(args, " ")))
			return true
		},
	}

	return Menu {
		previous: nil,
		entries: entries,
		getHelpDescription: func (gameState *GameState, player *Player, playerId int) string {
			return "You are in the world."
		},
	}
}
