package game

import (
	"fmt"
	"context"
	"time"
	"log"
	"mud/world"
)

const GAME_UPDATE_INTERVAL = world.WORLD_SECONDS_PER_UPDATE * time.Second

type Command struct {
	PlayerId int
	Payload string
}

type GameState struct {
	Commands chan Command

	players []Player
	playerIdToIndexMap map[int]int

	world *world.World
}

func GameStateInit() *GameState {
	playerMenusInit()
	world := world.WorldInitNew()

	gamestate := &GameState {
		Commands: make(chan Command, 1024),

		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),

		world: world,
	}

	return gamestate
}

func (gamestate *GameState) Run(ctx context.Context) {
	ticker := time.NewTicker(GAME_UPDATE_INTERVAL)
	defer ticker.Stop()

	gameloop:
	for {
		select {
			case <- ctx.Done():
				break gameloop
			case command := <- gamestate.Commands:
				gamestate.handleCommand(command)
			case <- ticker.C:
				gamestate.update()
		}
	}

	log.Printf("Shutdown signal received. Shutting down server...")
	// gamestate.world.Save("./world.json")
}

func (gamestate *GameState) RegisterPlayer(playerId int, playerInbox *chan string) {
	player := playerInit(playerId, playerInbox)
	gamestate.players = append(gamestate.players, player)
	newPlayerIndex := len(gamestate.players) - 1
	gamestate.playerIdToIndexMap[playerId] = newPlayerIndex

	newPlayer := &gamestate.players[newPlayerIndex]
	*newPlayer.inbox <- "Welcome to the RC Disco MUD!"
	// newPlayer.enterMenu(gamestate, &gamestate.menuLogin)
}

func (gamestate *GameState) RemovePlayer(playerId int) {
	// Get the player index (and double-check that they even exist)
	playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
	if !exists {
		log.Printf("Tried to remove player %d, but they don't exist!", playerId)
		return
	}

	// Check if they are logged in
	// player := &gamestate.players[playerIndex]
	// if player.isLoggedIn {
		// player.exitWorld(gamestate)
		// }

	// Swap and pop them from the array
	lastIndex := len(gamestate.players) - 1
	gamestate.players[playerIndex] = gamestate.players[lastIndex]
	gamestate.players = gamestate.players[:lastIndex]

	// Delete their entry in the map
	delete(gamestate.playerIdToIndexMap, playerId)

	// Tell everybody about it
	gamestate.broadcast(fmt.Sprintf("Player %d has left the game.", playerId))
}

func (gamestate *GameState) getPlayerById(playerId int) *Player {
	playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
	if !exists {
		return nil
	}

	return &gamestate.players[playerIndex]
}

// Handles a player command
func (gamestate *GameState) handleCommand(command Command) {
	// Lookup player index
	playerIndex, playerIndexExists := gamestate.playerIdToIndexMap[command.PlayerId]
	if !playerIndexExists {
		log.Printf("Received command from player %d but they don't exist.", command.PlayerId)
		return
	}

	player := &gamestate.players[playerIndex]
	player.getMenu().handleCommand(gamestate, player, command.Payload)
}

// Sends a message to all player inboxes
func (gamestate *GameState) broadcast(message string) {
	for _, player := range gamestate.players {
		*player.inbox <- message
	}
}

// This function is the update that is called on a 3-second interval
func (gamestate *GameState) update() {
	gamestate.world.Update()

	// Pass messages from the world update to the players
	for index := range len(gamestate.world.Messages) {
		message := &gamestate.world.Messages[index]
		for _, playerId := range message.ToPlayers {
			playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
			if !exists {
				continue
			}

			player := &gamestate.players[playerIndex]
			*player.inbox <- message.Message
		}
	}

	// Clear the messages
	clear(gamestate.world.Messages)
}
