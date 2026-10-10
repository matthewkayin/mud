package game

import (
	"context"
	"time"
	"log"
	"mud/util"
	"mud/world"
)

const GAME_UPDATE_INTERVAL = world.WORLD_SECONDS_PER_UPDATE * time.Second
const GAME_SAVE_INTERVAL = 15 * time.Minute

type SocketEventType int
const (
	SOCKET_EVENT_TYPE_CONNECT SocketEventType = iota
	SOCKET_EVENT_TYPE_DISCONNECT
	SOCKET_EVENT_TYPE_COMMAND
)

type SocketEvent struct {
	Type SocketEventType
	PlayerId int
	// Identifies the connection for CONNECT and DISCONNECT events
	Inbox *chan string
	// Only used by COMMAND events
	Command string
}

type EventListener func (gamestate *GameState, event* world.Event)

type GameState struct {
	SocketEvents chan SocketEvent
	bannerLines []string

	players []Player
	playerIdToIndexMap map[int]int

	world *world.World
	eventListeners [][]EventListener
}

func GameStateInit() *GameState {
	playerMenusInit()

	bannerLines, err := util.ReadFileLines("./banner.txt")
	if err != nil {
		log.Fatalf("Error loading ASCII banner: %s", err.Error())
	}

	gamestate := &GameState {
		SocketEvents: make(chan SocketEvent, 1024),
		bannerLines: bannerLines,

		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),

		world: world.WorldInit(),
		eventListeners: make([][]EventListener, world.EVENT_TYPE_COUNT),
	}

	gamestate.addEventListener(world.EVENT_TYPE_MESSAGE, handleEventMessage)
	gamestate.addEventListener(world.EVENT_TYPE_MOB_MOVE, tradeSessionOnMobMove)
	gamestate.addEventListener(world.EVENT_TYPE_MOB_DEATH, tradeSessionOnMobDeath)
	gamestate.addEventListener(world.EVENT_TYPE_MOB_DEATH, playerOnMobDeath)

	return gamestate
}

func (gamestate *GameState) Run(ctx context.Context) {
	ticker := time.NewTicker(GAME_UPDATE_INTERVAL)
	defer ticker.Stop()

	saveTicker := time.NewTicker(GAME_SAVE_INTERVAL)
	defer saveTicker.Stop()

	gameloop:
	for {
		select {
			case <- ctx.Done():
				break gameloop
			case event := <- gamestate.SocketEvents:
				switch event.Type {
					case SOCKET_EVENT_TYPE_CONNECT:
						gamestate.registerPlayer(event.PlayerId, event.Inbox)
					case SOCKET_EVENT_TYPE_DISCONNECT:
						gamestate.removePlayer(event.PlayerId, event.Inbox)
					case SOCKET_EVENT_TYPE_COMMAND:
						gamestate.handleCommand(event.PlayerId, event.Command)
					default:
						log.Printf("Socket event type %d not handled!", event.Type)
				}
			case <- ticker.C:
				gamestate.update()
			case <- saveTicker.C:
				gamestate.saveAllLoggedInPlayers()
		}
	}

	log.Printf("Shutdown signal received. Shutting down server...")
	gamestate.saveAllLoggedInPlayers()
	gamestate.world.ScriptQuit()
}

// Must only be called from the game loop
func (gamestate *GameState) registerPlayer(playerId int, playerInbox *chan string) {
	// If this player is already connected, kick the old connection
	existingPlayer := gamestate.getPlayerById(playerId)
	if existingPlayer != nil && existingPlayer.inbox != playerInbox {
		log.Printf("Player %d connected from another location. Kicking previous connection.", playerId)
		*existingPlayer.inbox <- "You have been disconnected because you logged in from another location."
		gamestate.removePlayer(playerId, existingPlayer.inbox)
	}

	player := playerInit(playerId, playerInbox)
	gamestate.players = append(gamestate.players, player)
	newPlayerIndex := len(gamestate.players) - 1
	gamestate.playerIdToIndexMap[playerId] = newPlayerIndex

	newPlayer := &gamestate.players[newPlayerIndex]
	for index := range len(gamestate.bannerLines) {
		*newPlayer.inbox <- gamestate.bannerLines[index]
	}
	newPlayer.setMenu(gamestate, PLAYER_MENU_LOGIN)
}

// Must only be called from the game loop
// The inbox identifies which connection is being removed. If it doesn't match the
// registered player's inbox (e.g. a kicked connection disconnecting), this does nothing.
func (gamestate *GameState) removePlayer(playerId int, playerInbox *chan string) {
	// Get the player index (and double-check that they even exist)
	playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
	if !exists {
		log.Printf("Warn - Tried to remove player %d, but they don't exist!", playerId)
		return
	}

	player := &gamestate.players[playerIndex]
	if player.inbox != playerInbox {
		return
	}

	// Check if they are logged in
	if player.isLoggedIn() {
		// A little hacky, the setMenu() will trigger the world menu on exit
		player.setMenu(gamestate, PLAYER_MENU_LOGIN)
	}

	// Closing the inbox tells the connection's write loop that the session has ended
	close(*player.inbox)

	// Swap and pop them from the array
	lastIndex := len(gamestate.players) - 1
	gamestate.players[playerIndex] = gamestate.players[lastIndex]
	gamestate.players = gamestate.players[:lastIndex]
	if playerIndex != lastIndex {
		gamestate.playerIdToIndexMap[gamestate.players[playerIndex].id] = playerIndex
	}

	// Delete their entry in the map
	delete(gamestate.playerIdToIndexMap, playerId)
}

func (gamestate *GameState) getPlayerById(playerId int) *Player {
	playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
	if !exists {
		return nil
	}

	return &gamestate.players[playerIndex]
}

func (gamestate *GameState) getPlayerByMobHandle(handle world.MobHandle) *Player {
	mob := gamestate.world.Mobs.Get(handle)
	if !mob.IsPlayer() {
		return nil
	}

	playerIndex, exists := gamestate.playerIdToIndexMap[mob.PlayerCharacter.PlayerId]
	if !exists {
		log.Printf("Warn - Tried get player %d by mob handle %d:%d, but they don't exist.",
			mob.PlayerCharacter.PlayerId, handle.Id, handle.Generation)
		return nil
	}

	return &gamestate.players[playerIndex]
}

func (gamestate *GameState) saveAllLoggedInPlayers() {
	for index := range len(gamestate.players) {
		player := &gamestate.players[index]
		if !player.isLoggedIn() {
			continue
		}

		world.SaveCharacter(player.character)
	}

	log.Printf("Saved all logged in player characters.")
}

// Handles a player command
func (gamestate *GameState) handleCommand(playerId int, command string) {
	// Lookup player index
	playerIndex, playerIndexExists := gamestate.playerIdToIndexMap[playerId]
	if !playerIndexExists {
		log.Printf("Received command from player %d but they don't exist.", playerId)
		return
	}

	player := &gamestate.players[playerIndex]
	player.getMenu().handleCommand(gamestate, player, command)
}

// Sends a message to all player inboxes
func (gamestate *GameState) broadcast(message string) {
	for _, player := range gamestate.players {
		*player.inbox <- message
	}
}

func (gamestate *GameState) messageRoom(roomIndex int, message string) {
	room := &gamestate.world.Rooms[roomIndex]

	for _, mobHandle := range room.Occupants {
		mob := gamestate.world.Mobs.Get(mobHandle)
		if !mob.IsPlayer() {
			continue
		}

		playerId := mob.PlayerCharacter.PlayerId
		playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
		if !exists {
			log.Printf("Warn - Mob %d:%d in room %d has player ID %d but that player does not exist.",
				mobHandle.Id, mobHandle.Generation, roomIndex, playerId)
			continue
		}

		player := &gamestate.players[playerIndex]
		*player.inbox <- message
	}
}

func (gamestate *GameState) addEventListener(eventType world.EventType, listener EventListener) {
	gamestate.eventListeners[eventType] = append(gamestate.eventListeners[eventType], listener)
}

func (gamestate *GameState) createCharacter(playerId int, characterSheet *world.CharacterSheet) {
	character := world.CharacterInitEmpty(playerId, characterSheet)

	// TEMP DEBUG DELETE ME
	character.SpellsKnown = append(character.SpellsKnown, 0)

	gamestate.world.AddCharacter(playerId, character)
	world.SaveCharacter(character)
}

// This function is the update that is called on a 3-second interval
func (gamestate *GameState) update() {
	// Handle player actions
	for index := 0; index < len(gamestate.players); index++ {
		if !gamestate.players[index].isLoggedIn() {
			continue
		}

		gamestate.players[index].doAction(gamestate)
	}

	// World update
	gamestate.world.Update()

	// Handle world events
	for index := range len(gamestate.world.Events) {
		event := &gamestate.world.Events[index]
		listeners := &gamestate.eventListeners[event.EventType]
		for _, listener := range *listeners {
			listener(gamestate, event)
		}
	}

	// Clear world events
	// This syntax for clearing the array is done because it keeps the underlying array
	// so that way we're not allocating a new chunk of memory each time we reset the events
	gamestate.world.Events = gamestate.world.Events[:0]
}

func handleEventMessage(gamestate *GameState, event* world.Event) {
	eventData := event.Data.(world.EventMessage)

	for _, playerId := range eventData.ToPlayers {
		playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
		if !exists {
			continue
		}

		player := &gamestate.players[playerIndex]
		*player.inbox <- eventData.Message
	}
}
