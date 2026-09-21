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
	entries: map[string]MenuEntry {
	},
}
