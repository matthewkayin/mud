package game

import (
	"fmt"
	"context"
	"time"
	"log"
	"strings"
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

func (gameState *GameState) RegisterPlayer(playerId int, playerInbox *chan string) {
	gameState.broadcast(fmt.Sprintf("Player %d has joined the game.\n", playerId))

	gameState.playerInboxes[playerId] = playerInbox
	*playerInbox <- "Welcome to the RC Disco MUD!\n"
}

func (gameState *GameState) RemovePlayer(playerId int) {
	delete(gameState.playerInboxes, playerId)

	gameState.broadcast(fmt.Sprintf("Player %d has left the game.\n", playerId))
}

// Handles a player command
func (gameState *GameState) handleCommand(command Command) {
	log.Printf("Received command. Player %d Payload %s", command.PlayerId, command.Payload)

	playerInbox, playerInboxExists := gameState.playerInboxes[command.PlayerId]
	if !playerInboxExists {
		log.Printf("No Inbox for player %d", command.PlayerId)
		return
	}

	words := strings.Split(command.Payload, " ")
	wordsLength := len(words)

	verb := words[0]
	
	switch verb {
		//case "quit", "logout":

		case "say":
			if wordsLength < 2 {
				*playerInbox <- "You must include a message that you want to say. For example: \"Say hello!\" will say \"hello!\""
				return
			}
			gameState.broadcast(fmt.Sprintf("Player %d: \"%s\"", command.PlayerId, strings.Join(words[1:], " ")))
			
		default:
			*playerInbox <- fmt.Sprintf("%s is not a legal action", verb)

	}

}

// Sends a message to all player inboxes
func (gameState *GameState) broadcast(message string) {
	for _, inbox := range gameState.playerInboxes {
		*inbox <- message
	}
}

func (gameState *GameState) update() {

}
