package game

import (
	"fmt"
	"mud/world"
)

type PlayerMenu int
const (
	PLAYER_MENU_LOGIN = iota
)

type Player struct {
	key string
	inbox *chan SocketMessage

	menu *Menu
	submenu Submenu

	character *world.Character
	mobHandle world.MobHandle
}

func playerInit(key string, inbox *chan SocketMessage) Player {
	return Player {
		key: key,
		inbox: inbox,
		menu: nil,
		submenu: nil,
		character: nil,
	}
}

func (player *Player) isLoggedIn() bool {
	return player.character != nil
}

func (player *Player) send(message string) {
	*player.inbox <- SocketMessage { Message: &message }
}

func (player *Player) sendHeartbeat(gamestate *GameState) {
	heartbeat := SocketMessage { Message: nil }

	// Send the heartbeat, but if the player's inbox is full, don't send it
	// This prevents a stalled connection from blocking the game loop
	select {
		case *player.inbox <- heartbeat:
		default:
	}
}

func (player *Player) setMenuLogin(gamestate *GameState) {
	player.menu = &gamestate.menuLogin
	player.send("You are in the login screen.")
	player.send("Type 'help' to see a list of options.")
}

func (player *Player) setMenuWorld(gamestate *GameState) {
	// Send a message to the room announcing the player's login
	gamestate.world.MessageRoom(player.character.Room, fmt.Sprintf("<y>%s</y> has joined the room.", player.character.Name))

	// Spawn player mob
	gamestate.spawnPlayerMob(player)

	// Enter world menu
	player.menu = &gamestate.menuWorld
	player.send(fmt.Sprintf("You have logged in. Welcome, <y>%s</y>.", player.character.Name))
	room := &gamestate.world.Rooms[player.character.Room]
	describeRoomToPlayer(gamestate, player, room)
}
