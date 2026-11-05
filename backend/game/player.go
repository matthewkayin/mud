package game

type PlayerMenu int
const (
	PLAYER_MENU_LOGIN = iota
)

type Player struct {
	key string
	inbox *chan SocketMessage
	menu *Menu
}

func playerInit(key string, inbox *chan SocketMessage) Player {
	return Player {
		key: key,
		inbox: inbox,
		menu: &MENU_LOGIN,
	}
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
