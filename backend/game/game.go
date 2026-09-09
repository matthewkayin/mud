package game

import (
	"fmt"
	"context"
	"time"
	"log"
)

const GAME_UPDATE_INTERVAL = 3 * time.Second

type PlayerMode int
const (
	PlayerModeMenuLogin PlayerMode = iota
	PlayerModeMenuCreateCharacter
	PlayerModeInGame
)

type Command struct {
	PlayerId int
	Payload string
}

type GameState struct {
	Commands chan Command

	// Menus
	menuLogin Menu
	menuCreateCharacter Menu
	menuWorld Menu

	// Players
	players []Player
	playerIdToIndexMap map[int]int

	world World
}

func InitState() *GameState {
	// Create menus
	menuLogin := MenuLogin()
	menuCreateCharacter := MenuCreateCharacter()
	menuWorld := MenuWorld()

	return &GameState {
		Commands: make(chan Command, 1024),

		menuLogin: menuLogin,
		menuCreateCharacter: menuCreateCharacter,
		menuWorld: menuWorld,

		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),

		world: WorldInit(),
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
	gameState.players = append(gameState.players, Player {
		id: playerId,
		inbox: playerInbox,
		menuInstance: nil,

		character: nil,
	})
	newPlayerIndex := len(gameState.players) - 1
	gameState.playerIdToIndexMap[playerId] = newPlayerIndex

	newPlayer := &gameState.players[newPlayerIndex]
	*newPlayer.inbox <- "Welcome to the RC Disco MUD!"
	newPlayer.enterMenu(gameState, &gameState.menuLogin)
}

func (gameState *GameState) RemovePlayer(playerId int) {
	// Get the player index (and double-check that they even exist)
	playerIndex, exists := gameState.playerIdToIndexMap[playerId]
	if !exists {
		log.Printf("Tried to remove player %d, but they don't exist!", playerId)
		return
	}

	// Swap and pop them from the array
	lastIndex := len(gameState.players) - 1
	gameState.players[playerIndex] = gameState.players[lastIndex]
	gameState.players = gameState.players[:lastIndex]

	// Delete their entry in the map
	delete(gameState.playerIdToIndexMap, playerId)

	// Tell everybody about it
	gameState.broadcast(fmt.Sprintf("Player %d has left the game.", playerId))
}

// Handles a player command
func (gameState *GameState) handleCommand(command Command) {
	log.Printf("Received command. Player %d Payload %s", command.PlayerId, command.Payload)

	// Lookup player index
	playerIndex, playerIndexExists := gameState.playerIdToIndexMap[command.PlayerId]
	if !playerIndexExists {
		log.Printf("Received command from player %d but they don't exist.", command.PlayerId)
		return
	}

	// Handle command using the player's current menu
	player := &gameState.players[playerIndex]
	if player.menuInstance == nil {
		log.Printf("Received command from player %d but they don't have a menu instance.", command.PlayerId)

		// Remove the player because they are in an unrecoverable state
		// TODO: We should also develop a way to kick them / i.e. trigger a close in their web socket connection
		gameState.RemovePlayer(command.PlayerId)
		return
	}

	player.menuInstance.HandleCommand(gameState, player, command.Payload)
}

// Sends a message to all player inboxes
func (gameState *GameState) broadcast(message string) {
	for _, player := range gameState.players {
		*player.inbox <- message
	}
}

func (gameState *GameState) update() {

}
