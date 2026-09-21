package game

type PlayerMode int
const (
	PLAYER_MODE_MENU_LOGIN = iota
	PLAYER_MODE_MENU_CREATE
	PLAYER_MODE_IN_WORLD
)

type Player struct {
	id int
	inbox *chan string
	mode PlayerMode

	// nextAction Action
	// character *Character
	// mobHandle MobHandle
	// tradeSession *TradeSession
}

func playerInit(id int, inbox *chan string) Player {
	return Player {
		id: id,
		inbox: inbox,
		mode: PLAYER_MODE_MENU_LOGIN,
	}
}
