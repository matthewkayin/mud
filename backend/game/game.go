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

type Player struct {
	id int
	inbox *chan string
	menu *Menu

	character *Character
}

type GameState struct {
	Commands chan Command

	menuLogin Menu
	menuCreateCharacter Menu
	menuWorld Menu

	players []Player
	playerIdToIndexMap map[int]int

	world World
}

func InitState() *GameState {
	// Create menus
	menuLogin := MenuLogin()
	menuCreateCharacter := MenuCreateCharacter(&menuLogin)
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
	gameState.broadcast(fmt.Sprintf("Player %d has joined the game.", playerId))

	gameState.players = append(gameState.players, Player {
		id: playerId,
		inbox: playerInbox,
		menu: &gameState.menuLogin,

		character: nil,
	})
	newPlayerIndex := len(gameState.players) - 1
	gameState.playerIdToIndexMap[playerId] = newPlayerIndex

	*playerInbox <- "Welcome to the RC Disco MUD!"
	*playerInbox <- "Type \"login <character>\" to login or \"help\" for more options."
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
	if player.menu == nil {
		log.Printf("Received command from player %d but they don't have a menu.", command.PlayerId)

		// Remove the player because they are in an unrecoverable state
		// TODO: We should also develop a way to kick them / i.e. trigger a close in their web socket connection
		gameState.RemovePlayer(command.PlayerId)
		return
	}

	player.menu.HandleCommand(gameState, player, command.Payload)
}

func (gameState *GameState) setPlayerMenu(player *Player, menu *Menu) {
	player.menu = menu
	if player.menu.onEnter != nil {
		player.menu.onEnter(gameState, player)
	}
}

// Sends a message to all player inboxes
func (gameState *GameState) broadcast(message string) {
	for _, player := range gameState.players {
		*(player.inbox) <- message
	}
}

func (gameState *GameState) update() {

}
