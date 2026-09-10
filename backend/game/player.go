package game

import (
	"log"
	"fmt"
)

type Player struct {
	id int
	inbox *chan string
	isLoggedIn bool
	menuInstance *MenuInstance
	character *Character
	mobHandle MobHandle
}

func PlayerInit(playerId int, playerInbox *chan string) Player {
	return Player {
		id: playerId,
		inbox: playerInbox,
		isLoggedIn: false,
		menuInstance: nil,
		character: nil,
	}
}

func (player *Player) exitMenu(gameState* GameState) {
	if player.menuInstance == nil {
		log.Printf("Player %d has no menu!", player.id)
		return
	}

	if player.menuInstance.previous == nil {
		log.Printf("Player %d tried to go back in a menu, but they have no previous menu instance.", player.id)
		return
	}

	player.menuInstance = player.menuInstance.previous
	player.menuInstance.menu.onEnter(gameState, player)
}

func (player *Player) enterMenu(gameState *GameState, menu *Menu) {
	previous := player.menuInstance
	player.menuInstance = menu.createInstance()
	player.menuInstance.previous = previous

	*player.inbox <- player.menuInstance.menu.getDescription(gameState, player)
	if player.menuInstance.menu.onEnter != nil {
		player.menuInstance.menu.onEnter(gameState, player)
	}
}

func (player *Player) enterWorld(gameState *GameState, asCharacter *Character) {
	player.isLoggedIn = true
	player.character = asCharacter

	// Create a mob for the player
	playerMob := MobInitFromCharacter(player.character)
	player.mobHandle = gameState.world.Mobs.Push(playerMob)
	playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
	playerRoom.AddOccupant(player.mobHandle)

	// Enter world menu
	*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", player.character.Data.Name)
	player.enterMenu(gameState, &gameState.menuWorld)
}

func (player *Player) exitWorld(gameState *GameState) {
	// Remove player from current room
	playerMob := gameState.world.Mobs.Get(player.mobHandle)
	playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
	playerRoom.RemoveOccupant(player.mobHandle)

	// Save player mob data back to their character
	player.character.Data = playerMob.Data

	player.isLoggedIn = false
	player.enterMenu(gameState, &gameState.menuLogin)
}
