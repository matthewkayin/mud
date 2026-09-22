package game

import (
	"fmt"
	"mud/world"
)

var MENU_WORLD = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		// Clear the player's action in case they had any leftover from a previous login session
		player.nextAction = Action {
			actionType: ACTION_TYPE_NONE,
			data: nil,
		}

		// Create a mob for the player
		playerMob := world.MobInitFromCharacter(player.character)
		player.mobHandle = gamestate.world.Mobs.Push(playerMob)
		playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
		gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s has joined the room.", player.character.Data.Name))
		playerRoom.AddOccupant(player.mobHandle)

		// Enter world menu
		*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", player.character.Data.Name)
		*player.inbox <- fmt.Sprintf("You are in %s.", playerRoom.Name)
	},
	onExit: func(gamestate *GameState, player *Player) {
		// Remove player from current room
		playerMob := gamestate.world.Mobs.Get(player.mobHandle)
		playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
		playerRoom.RemoveOccupant(player.mobHandle)
		gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("%s has left the world.", playerMob.Data.Name))

		// Save player mob data back to their character
		player.character.Data = playerMob.Data
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
