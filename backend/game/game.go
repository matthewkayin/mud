package game

import (
	"fmt"
	"context"
	"time"
	"log"
)

const GAME_SECONDS_PER_UPDATE = 3
const GAME_UPDATE_INTERVAL = GAME_SECONDS_PER_UPDATE * time.Second

// Each entry in this array corresponds to a player mode
var MENUS []*Menu = []*Menu {
	&MENU_LOGIN,
	&MENU_CHARACTER_CREATOR,
	&MENU_WORLD,
}

type Command struct {
	PlayerId int
	Payload string
}

type GameState struct {
	Commands chan Command

	players []Player
	playerIdToIndexMap map[int]int
}

func GameStateInit() *GameState {
	gameState := &GameState {
		Commands: make(chan Command, 1024),

		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),
	}

	return gameState
}

func (gameState *GameState) Run(ctx context.Context) {
	ticker := time.NewTicker(GAME_UPDATE_INTERVAL)
	defer ticker.Stop()

	gameloop:
	for {
		select {
			case <- ctx.Done():
				break gameloop
			case command := <- gameState.Commands:
				gameState.handleCommand(command)
			case <- ticker.C:
				gameState.update()
		}
	}

	log.Printf("Shutdown signal received. Shutting down server...")
	// gameState.world.Save("./world.json")
}

func (gameState *GameState) RegisterPlayer(playerId int, playerInbox *chan string) {
	player := playerInit(playerId, playerInbox)
	gameState.players = append(gameState.players, player)
	newPlayerIndex := len(gameState.players) - 1
	gameState.playerIdToIndexMap[playerId] = newPlayerIndex

	newPlayer := &gameState.players[newPlayerIndex]
	*newPlayer.inbox <- "Welcome to the RC Disco MUD!"
	// newPlayer.enterMenu(gameState, &gameState.menuLogin)
}

func (gameState *GameState) RemovePlayer(playerId int) {
	// Get the player index (and double-check that they even exist)
	playerIndex, exists := gameState.playerIdToIndexMap[playerId]
	if !exists {
		log.Printf("Tried to remove player %d, but they don't exist!", playerId)
		return
	}

	// Check if they are logged in
	// player := &gameState.players[playerIndex]
	// if player.isLoggedIn {
		// player.exitWorld(gameState)
		// }

	// Swap and pop them from the array
	lastIndex := len(gameState.players) - 1
	gameState.players[playerIndex] = gameState.players[lastIndex]
	gameState.players = gameState.players[:lastIndex]

	// Delete their entry in the map
	delete(gameState.playerIdToIndexMap, playerId)

	// Tell everybody about it
	gameState.broadcast(fmt.Sprintf("Player %d has left the game.", playerId))
}

func (gameState *GameState) getPlayerById(playerId int) *Player {
	playerIndex, exists := gameState.playerIdToIndexMap[playerId]
	if !exists {
		return nil
	}

	return &gameState.players[playerIndex]
}

// Handles a player command
func (gameState *GameState) handleCommand(command Command) {
	// Lookup player index
	playerIndex, playerIndexExists := gameState.playerIdToIndexMap[command.PlayerId]
	if !playerIndexExists {
		log.Printf("Received command from player %d but they don't exist.", command.PlayerId)
		return
	}

	player := &gameState.players[playerIndex]
	menu := MENUS[player.mode]
	menu.handleCommand(gameState, player, command.Payload)
}

// Sends a message to all player inboxes
func (gameState *GameState) broadcast(message string) {
	for _, player := range gameState.players {
		*player.inbox <- message
	}
}

// This function is the update that is called on a 3-second interval
func (gameState *GameState) update() {
}
