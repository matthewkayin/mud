package game

import (
	"context"
	"log"
	"mud/util"
	"mud/world"
	"time"
)

const GAME_UPDATE_INTERVAL = world.WORLD_SECONDS_PER_UPDATE * time.Second
const GAME_SAVE_INTERVAL = 5 * time.Minute

type ServerMode int
const (
	SERVER_MODE_RUNNING ServerMode = iota
	SERVER_MODE_SHUTDOWN_SCHEDULED
	SERVER_MODE_MAINTENANCE
)

type GameState struct {
	serverMode ServerMode
	version string
	welcomeBannerLines []string

	menuLogin Menu
	menuWorld Menu

	SocketEvents chan SocketEvent

	players []Player
	playerKeyToIndexMap map[string]int
	// A list of logged in player names, which is cached here so that we don't have to re-allocate it every 'tell'
	playerCharacterNames []string
	// A list of logged in player keys. This list is parallel to playerCharacterNames and is used so that we know which player to 'tell' to
	playerCharacterKeys []string

	world *world.World
}

func Init(version string) *GameState {
	// Load welcome banner
	welcomeBannerLines, err := util.ReadFileLines("./banner.txt")
	if err != nil {
		log.Fatalf("Error loading welcome banner: %s", err.Error())
	}

	// Load world
	newWorld, err := world.NewWorld()
	if err != nil {
		log.Fatalf("Failed to load world: %s", err.Error())
	}

	gamestate := &GameState {
		serverMode: SERVER_MODE_RUNNING,
		version: version,
		welcomeBannerLines: welcomeBannerLines,

		SocketEvents: make(chan SocketEvent, 1024),

		players: make([]Player, 0, 64),
		playerKeyToIndexMap: make(map[string]int),
		playerCharacterNames: make([]string, 0, 64),
		playerCharacterKeys: make([]string, 0, 64),

		world: newWorld,
	}

	gamestate.world.MessagePlayer = gamestate.messagePlayer

	gamestate.initMenus()

	return gamestate
}

func (gamestate *GameState) initMenus() {
	gamestate.menuLogin = NewMenu(MENU_LOGIN_PARAMS)
	gamestate.menuWorld = NewMenu(MENU_WORLD_PARAMS)
}

// RUN

func (gamestate *GameState) Run(ctx context.Context) {
	ticker := time.NewTicker(GAME_UPDATE_INTERVAL)
	defer ticker.Stop()

	saveTicker := time.NewTicker(GAME_SAVE_INTERVAL)
	defer saveTicker.Stop()

	gameloop:
	for {
		select {
			case <- ctx.Done(): {
				break gameloop
			}

			case event := <- gamestate.SocketEvents: {
				switch event.Type {
					case SOCKET_EVENT_TYPE_CONNECT: {
						gamestate.registerPlayer(event.PlayerKey, event.Inbox)
					}

					case SOCKET_EVENT_TYPE_DISCONNECT: {
						gamestate.removePlayer(event.PlayerKey, event.Inbox)
					}

					case SOCKET_EVENT_TYPE_COMMAND: {
						gamestate.handleCommand(event.PlayerKey, event.Command)
					}

					default: {
						log.Printf("Warn: Socket event type %d not handled!", event.Type)
					}
				}
			}

			case <- ticker.C: {
				gamestate.update()
			}

			case <- saveTicker.C: {
				gamestate.savePlayers()
			}
		}
	}

	log.Printf("Shutdown signal received. Shutting down server...")
	gamestate.savePlayers()
	gamestate.world.Quit()
}

// PLAYERS

func (gamestate *GameState) registerPlayer(playerKey string, playerInbox *chan SocketMessage) {
	// Refuse connections during maintenance
	if gamestate.serverMode == SERVER_MODE_MAINTENANCE {
		message := "The server is down for maintenance. Please try again later."
		*playerInbox <- SocketMessage { Message: &message }
		close(*playerInbox)
		return
	}

	// If this player is already connected, kick the old connection
	existingPlayer := gamestate.getPlayerByKey(playerKey)
	if existingPlayer != nil && existingPlayer.inbox != playerInbox {
		log.Printf("Player %s connected from another location. Kicking previous connection.", playerKey)
		existingPlayer.send("You have been disconnected because you logged in from another location.")
		gamestate.removePlayer(playerKey, existingPlayer.inbox)
	}

	// Init player and add them to the player list
	player := playerInit(playerKey, playerInbox)
	playerIndex := len(gamestate.players)
	gamestate.players = append(gamestate.players, player)
	gamestate.playerKeyToIndexMap[playerKey] = playerIndex

	// Send welcome banner
	playerPtr := &gamestate.players[playerIndex]
	for index := range len(gamestate.welcomeBannerLines) {
		playerPtr.send(gamestate.welcomeBannerLines[index])
	}

	// Set them to the login menu
	playerPtr.setMenuLogin(gamestate)
}

func (gamestate *GameState) removePlayer(playerKey string, playerInbox *chan SocketMessage) {
	// Get the player index (and double-check that they event exist)
	playerIndex, exists := gamestate.playerKeyToIndexMap[playerKey]
	if !exists {
		log.Printf("Warn: Tried to remove player %s, but they don't exist.", playerKey)
		return
	}

	player := &gamestate.players[playerIndex]
	if player.inbox != playerInbox {
		return
	}

	// Logout the player if they are logged in
	if player.isLoggedIn() {
		player.logout(gamestate)
	}

	close(*player.inbox)

	// Swap and pop them from the array
	lastIndex := len(gamestate.players) - 1
	gamestate.players[playerIndex] = gamestate.players[lastIndex]
	gamestate.players = gamestate.players[:lastIndex]
	if playerIndex != lastIndex {
		gamestate.playerKeyToIndexMap[gamestate.players[playerIndex].key] = playerIndex
	}

	// Delete their map entry
	delete(gamestate.playerKeyToIndexMap, playerKey)
}

func (gamestate *GameState) getPlayerByKey(playerKey string) *Player {
	playerIndex, exists := gamestate.playerKeyToIndexMap[playerKey]
	if !exists {
		return nil
	}

	return &gamestate.players[playerIndex]
}

func (gamestate *GameState) spawnPlayerMob(player *Player) {
	playerMob := world.NewMobFromCharacter(player.character)
	player.mobHandle = gamestate.world.Mobs.Push(playerMob)
	playerRoom := &gamestate.world.Rooms[playerMob.Room]
	playerRoom.AddOccupant(player.mobHandle)
	player.isMinimapDirty = true
	player.isInventoryDirty = true
}

func (gamestate *GameState) messagePlayer(playerKey string, message string) {
	player := gamestate.getPlayerByKey(playerKey)
	if player == nil {
		return
	}

	player.send(message)
}

// Sends a message to all logged in players
func (gamestate *GameState) messageWorld(message string) {
	for index := range len(gamestate.players) {
		player := &gamestate.players[index]
		if player.isLoggedIn() {
			player.send(message)
		}
	}
}

func (gamestate *GameState) savePlayers() {
	for index := range len(gamestate.players) {
		player := &gamestate.players[index]
		if !player.isLoggedIn() {
			continue
		}

		gamestate.world.SaveCharacter(player.character)
	}

	log.Printf("Saved all logged in player characters.")
}

// HANDLE COMMAND

func (gamestate *GameState) handleCommand(playerKey string, command string) {
	// Lookup player index
	player := gamestate.getPlayerByKey(playerKey)
	if player == nil {
		log.Printf("Warn: Received command from player %s, but they don't exist.", playerKey)
	}

	player.menu.handleCommand(gamestate, player, command)
	player.sendHeartbeat(gamestate)
}

// UPDATE

func (gamestate *GameState) update() {
	gamestate.world.Update()

	// Send heartbeats to players
	for index := range len(gamestate.players) {
		gamestate.players[index].sendHeartbeat(gamestate)
	}
}
