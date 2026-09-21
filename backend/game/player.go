package game

import (
	"mud/world"
)

type PlayerMenu int
const (
	PLAYER_MENU_LOGIN = iota
	PLAYER_MENU_CREATE
	PLAYER_MENU_WORLD
)

var PLAYER_MENUS []*Menu

type Player struct {
	id int
	inbox *chan string
	menu PlayerMenu

	// nextAction Action
	characterSheet *world.CharacterSheet
	character *world.Character
	// mobHandle MobHandle
	// tradeSession *TradeSession
}

func playerMenusInit() {
	PLAYER_MENUS = []*Menu {
		PLAYER_MENU_LOGIN: &MENU_LOGIN,
		PLAYER_MENU_CREATE: &MENU_CREATE,
		PLAYER_MENU_WORLD: &MENU_WORLD,
	}
}

func playerInit(id int, inbox *chan string) Player {
	return Player {
		id: id,
		inbox: inbox,
		menu: PLAYER_MENU_LOGIN,
		characterSheet: nil,
		character: nil,
	}
}

func (player *Player) isLoggedIn() bool {
	return player.character == nil
}

func (player *Player) getMenu() *Menu {
	return PLAYER_MENUS[player.menu]
}

func (player *Player) setMenu(gamestate *GameState, menu PlayerMenu) {
	player.getMenu().onExit(gamestate, player)
	player.menu = menu
	player.getMenu().onEnter(gamestate, player)
}
