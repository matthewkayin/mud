package game

import (
	"fmt"
	"context"
	"time"
	"log"
	"strings"
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
	mode PlayerMode
	inbox *chan string
}

type GameState struct {
	Commands chan Command

	players []Player
	playerIdToIndexMap map[int]int

	commandRegistry map[string]CommandRegistryEntry
}

func InitState() *GameState {
	return &GameState {
		Commands: make(chan Command, 1024),
		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),
		commandRegistry: CommandRegistryInit(),
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
		mode: PlayerModeMenuLogin,
		inbox: playerInbox,
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

	// Get pointer to the player
	player := &gameState.players[playerIndex]

	// Get command verb and arguments
	words := strings.Split(command.Payload, " ")
	verb := words[0]
	args := words[1:]

	// Lookup command in registry
	registryEntry, registryEntryExists := gameState.commandRegistry[verb]
	if !registryEntryExists {
		*(player.inbox) <- fmt.Sprintf("%s is not a legal action.", verb)
		return
	}

	// Execute command
	commandExecutedSuccessfully := registryEntry.handler(gameState, player, command.PlayerId, args)

	// Print command usage back to player
	if !commandExecutedSuccessfully {
		*(player.inbox) <- registryEntry.usage
	}
}

func (gameState *GameState) handleCommandHelp(player *Player) {
	switch player.mode {
		case PlayerModeMenuLogin:
			*(player.inbox) <- ""
		case PlayerModeMenuCreateCharacter:
		case PlayerModeInGame:
		default:
			log.Printf("Unhandled player mode %d.", player.mode)
	}
}

func (gameState *GameState) handleCommandSay(playerId int, player *Player, args []string) {
	if len(args) < 1 {
		*(player.inbox) <- "You must include a message that you want to say. For example: \"Say hello!\" will say \"hello!\""
		return
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
