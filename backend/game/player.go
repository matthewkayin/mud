package game

import (
	"log"
)

type Player struct {
	id int
	inbox *chan string
	menuInstance *MenuInstance
	character *Character
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

func (player *Player) enterMenu(gameState* GameState, menu *Menu) {
	previous := player.menuInstance
	player.menuInstance = menu.createInstance()
	player.menuInstance.previous = previous

	*player.inbox <- player.menuInstance.menu.getDescription(gameState, player)
	if player.menuInstance.menu.onEnter != nil {
		player.menuInstance.menu.onEnter(gameState, player)
	}
}
