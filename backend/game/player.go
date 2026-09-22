package game

import (
	"fmt"
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

	characterSheet *world.CharacterSheet
	character *world.Character
	nextAction Action
	mobHandle world.MobHandle
	tradeSession *TradeSession
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

func playerOnMobDeath(gamestate *GameState, event *world.Event) {
	eventData := event.Data.(world.EventMobDeath)

	if eventData.PlayerId == world.MOB_PLAYER_NONE {
		return
	}

	player := gamestate.getPlayerById(eventData.PlayerId)
	*player.inbox <- fmt.Sprintf("Your character %s has died, and death is forever. RIP", player.character.Data.Name)

	gamestate.world.RemoveCharacter(player.character)
	player.setMenu(gamestate, PLAYER_MENU_LOGIN)
}
