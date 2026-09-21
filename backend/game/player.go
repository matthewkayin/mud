package game

import (
	"mud/world"
)

type Player struct {
	id int
	inbox *chan string
	menu *Menu

	// nextAction Action
	character *world.Character
	// mobHandle MobHandle
	// tradeSession *TradeSession
}

func playerInit(id int, inbox *chan string) Player {
	return Player {
		id: id,
		inbox: inbox,
		menu: &MENU_LOGIN,
		character: nil,
	}
}

func (player *Player) isLoggedIn() bool {
	return player.character == nil
}

func (player *Player) setMenu(gamestate *GameState, menu *Menu) {
	player.menu = menu
	player.menu.onEnter(gamestate, player)
}
