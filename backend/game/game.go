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

	character *Character
}

type GameState struct {
	Commands chan Command
	commandRegistry map[string]CommandRegistryEntry

	players []Player
	playerIdToIndexMap map[int]int

	world World
}

func InitState() *GameState {
	return &GameState {
		Commands: make(chan Command, 1024),
		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),
		commandRegistry: CommandRegistryInit(),
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

	if player.mode == PlayerModeMenuCreateCharacter {
		// Handle back
		if command.Payload == "back" {
			player.mode = PlayerModeMenuLogin
			return
		}

		// Handle empty name
		if len(command.Payload) == 0 {
			*(player.inbox) <- "No name provided! Please enter a name or type back."
			return
		}

		// Handle name with space
		if strings.Contains(command.Payload, " ") {
			*(player.inbox) <- "Hey man you like, can't put a space in there."
			return
		}

		// Check if character exists
		_, characterExists := gameState.world.characters[command.Payload]
		if characterExists {
			*(player.inbox) <- fmt.Sprintf("A character named \"%s\" already exists.", command.Payload)
			return
		}

		// Create the character
		gameState.world.CreateCharacter(command.PlayerId, Character {
			playerId: command.PlayerId,
			name: command.Payload,
		})
		player.mode = PlayerModeMenuLogin
		*(player.inbox) <- fmt.Sprintf("Your character has been created. Type \"login %s\" to login to them.", command.Payload)

		return
	}

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

// Sends a message to all player inboxes
func (gameState *GameState) broadcast(message string) {
	for _, player := range gameState.players {
		*(player.inbox) <- message
	}
}

func (gameState *GameState) update() {

}
