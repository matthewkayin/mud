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
			gameState.setPlayerMenu(player, &gameState.menuLogin)
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
			if len(room.playersInRoom) > 1 {
				otherPlayerCount := len(room.playersInRoom) - 1
				otherPlayersCounted := 0
				otherPlayersStr := ""

				for _, otherPlayerId := range room.playersInRoom {
					// Don't tell the player about themselves being in the room
					if otherPlayerId == player.id {
						continue
					}

					// Get a handle to the other player
					otherPlayerIndex := gameState.playerIdToIndexMap[otherPlayerId]
					otherPlayer := &gameState.players[otherPlayerIndex]

					// Add their name to the string
					otherPlayersStr += otherPlayer.character.name

					// The format of other players string will be as follows:
					// 1 other player => <player name>
					// 2 other players => <player1> and <player2>
					// 3 or more => <player1>, <player2>, and <player3>

					// Note: we have to use otherPlayersCounted rather than index
					// because we don't know whether the current player will be the
					// last player in the list
					isLastOtherPlayer := otherPlayersCounted == otherPlayerCount - 1
					isSecondToLastOtherPlayer := otherPlayersCounted == otherPlayerCount - 2

					// Add comma and 'and' to the string as necessary
					if otherPlayerCount >= 3 && !isLastOtherPlayer {
						otherPlayersStr += ", "
					}
					if otherPlayerCount >= 2 && isSecondToLastOtherPlayer {
						otherPlayersStr += "and "
					}

					otherPlayersCounted += 1
				}

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
		previous: nil,
		entries: entries,
		getHelpDescription: func (gameState *GameState, player *Player) string {
			return "You are in the world."
		},
		onEnter: nil,
	}
}
