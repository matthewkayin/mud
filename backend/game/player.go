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
	inbox *chan SocketMessage
	menu PlayerMenu

	characterSheet *world.CharacterSheet
	character *world.Character
	nextAction Action
	mobHandle world.MobHandle
	tradeSession *TradeSession

	isTalking bool
	talkingToHandle world.MobHandle
}

func playerMenusInit() {
	PLAYER_MENUS = []*Menu {
		PLAYER_MENU_LOGIN: &MENU_LOGIN,
		PLAYER_MENU_CREATE: &MENU_CREATE,
		PLAYER_MENU_WORLD: &MENU_WORLD,
	}
}

func playerInit(id int, inbox *chan SocketMessage) Player {
	return Player {
		id: id,
		inbox: inbox,
		menu: PLAYER_MENU_LOGIN,
		characterSheet: nil,
		character: nil,
	}
}

func (player *Player) isLoggedIn() bool {
	return player.character != nil
}

func (player *Player) send(message string) {
	*player.inbox <- SocketMessage { Message: &message }
}

// Sends a message with no text. While the player is in-game, it carries their current status.
func (player *Player) sendHeartbeat(gamestate *GameState) {
	heartbeat := SocketMessage { Message: nil }
	if player.isLoggedIn() {
		playerMob, mobExists := gamestate.world.Mobs.GetIfExists(player.mobHandle)
		if mobExists {
			heartbeat.PlayerStatus = &PlayerStatus {
				Health: playerMob.Data.Health,
				MaxHealth: playerMob.Data.MaxHealth(),
				Mana: playerMob.Data.Mana,
				MaxMana: playerMob.Data.MaxMana(),
				PendingAction: player.nextAction.actionType != ACTION_TYPE_NONE,
			}
		}
	}

	// Send the heartbeat, but if the player's inbox is full, then don't send it
	// This prevents a stalled connection from blocking the game loop
	select {
		case *player.inbox <- heartbeat:
		default:
	}
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
	player.send(fmt.Sprintf("Your character %s has died, and death is forever. RIP", player.character.Data.Name))

	gamestate.world.RemoveCharacter(player.character)
	player.setMenu(gamestate, PLAYER_MENU_LOGIN)
}
