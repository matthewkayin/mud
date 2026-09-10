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
			// Remove player from current room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
			playerRoom.RemoveOccupant(player.mobHandle)

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
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]

			*(player.inbox) <- room.Description

			// Send the list of players in the room
			if len(room.Occupants) > 1 {
				otherPlayerCount := len(room.Occupants) - 1

				otherPlayerNames := make([]string, 0, otherPlayerCount)
				for _, mobHandle := range room.Occupants {
					// Don't tell the player about themselves being in the room
					if mobHandle.Equals(player.mobHandle) {
						continue
					}

					// Get a pointer to the mob
					mob := gameState.world.Mobs.Get(mobHandle)
					// Add their name to the list
					otherPlayerNames = append(otherPlayerNames, mob.Data.Name)
				}

				otherPlayersStr := menuWorldCombineNames(otherPlayerNames)
				*(player.inbox) <- fmt.Sprintf("%s are here.", otherPlayersStr)
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

			gameState.broadcast(fmt.Sprintf("%s: '%s'", player.character.Data.Name, strings.Join(args, " ")))
			return true
		},
	}

	return Menu {
		entries: entries,
		getDescription: func (gameState *GameState, player *Player) string {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]
			return fmt.Sprintf("You are in %s.", room.Name)
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
