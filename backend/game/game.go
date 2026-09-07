package game

import (
	"context"
	"time"
	"log"
)

const GAME_UPDATE_INTERVAL = 3 * time.Second

type Command struct {
	PlayerId int
	Payload string
}

type GameState struct {
	Commands chan Command
	playerInboxes map[int]*chan string
}

func InitState() *GameState {
	return &GameState {
		Commands: make(chan Command, 1024),
		playerInboxes: make(map[int]*chan string),
	}
}

func (gameState *GameState) Run(ctx context.Context) {
	ticker := time.NewTicker(GAME_UPDATE_INTERVAL)
	defer ticker.Stop()

	gameLoop:
	for {
		select {
			case <- ctx.Done():
				break gameLoop
			case command := <- gameState.Commands:
				gameState.handleCommand(command)
			case <- ticker.C:
				gameState.update()
		}
	}

	// TODO: end of game loop, save off game state data before exiting
}

func (gameState *GameState) RegisterPlayer(playerId int, outbox *chan string) {
	gameState.playerInboxes[playerId] = outbox
	*outbox <- "Welcome to the RC Disco MUD!"
}

func (gameState *GameState) handleCommand(command Command) {
	log.Printf("Received command. Player %d Payload %s", command.PlayerId, command.Payload)
}

func (gameState *GameState) update() {

}
