package game

import (
	"fmt"
)

var MENU_WORLD = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		// Clear the player's action in case they had any leftover from a previous login session
		// player.nextAction = Action {
			// actionType: ACTION_TYPE_NONE,
			// data: nil,
		// }

		// Create a mob for the player
		// playerMob := MobInitPlayer(player, player.character)
		// player.mobHandle = gameState.world.Mobs.Push(playerMob)
		// playerRoom := &gameState.world.Rooms[playerMob.data.Room]
		// playerRoom.broadcast(gameState, fmt.Sprintf("%s has joined the room.", player.character.Data.Name))
		// playerRoom.AddOccupant(player.mobHandle)

		// Enter world menu
		*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", player.character.Data.Name)
		// TODO: use the actual room name, and say who is here
		*player.inbox <- "You are in the Presentation Space."
	},
	onExit: func(gamestate *GameState, player *Player) {
		// Remove player mob and such
		// Fire event
		/*
		gameState.fireEvent(Event {
			eventType: EVENT_TYPE_PLAYER_LOGOUT,
			data: EventPlayerLogout {
				playerId: player.id,
			},
		})

		// Remove player from current room
		playerMob := gameState.world.Mobs.Get(player.mobHandle)
		playerRoom := &gameState.world.Rooms[playerMob.data.Room]
		playerRoom.RemoveOccupant(player.mobHandle)
		playerRoom.broadcast(gameState, fmt.Sprintf("%s has left the world.", playerMob.data.Name))

		// Save player mob data back to their character
		player.character.Data = playerMob.data
		*/
		player.character = nil
	},

	entries: map[string]MenuEntry {
		"logout": {
			usage: "logout",
			description: "Logout of the world",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)
				return true
			},
		},
	},
}
