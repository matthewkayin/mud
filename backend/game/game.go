package game

import (
	"fmt"
	"context"
	"time"
	"log"
	"mud/util"
	"mud/world"
)

const GAME_UPDATE_INTERVAL = world.WORLD_SECONDS_PER_UPDATE * time.Second
const GAME_SAVE_INTERVAL = 15 * time.Minute

type Command struct {
	PlayerId int
	Payload string
}

type EventListener func (gamestate *GameState, event* world.Event)

type GameState struct {
	Commands chan Command
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
		Commands: make(chan Command, 1024),
		bannerLines: bannerLines,

		players: make([]Player, 0, 64),
		playerIdToIndexMap: make(map[int]int),

		world: world.WorldInit(),
		eventListeners: make([][]EventListener, world.EVENT_TYPE_COUNT),
	}

	gamestate.addEventListener(world.EVENT_TYPE_MESSAGE, handleEventMessage)
	gamestate.addEventListener(world.EVENT_TYPE_MOB_MOVE, tradeSessionOnMobMove)
	gamestate.addEventListener(world.EVENT_TYPE_MOB_SET_TARGET,  tradeSessionOnMobSetTarget)
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
			case command := <- gamestate.Commands:
				gamestate.handleCommand(command)
			case <- ticker.C:
				gamestate.update()
			case <- saveTicker.C:
				gamestate.world.Save()
		}
	}

	log.Printf("Shutdown signal received. Shutting down server...")
	gamestate.world.Save()
}

func (gamestate *GameState) RegisterPlayer(playerId int, playerInbox *chan string) {
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

func (gamestate *GameState) RemovePlayer(playerId int) {
	// Get the player index (and double-check that they even exist)
	playerIndex, exists := gamestate.playerIdToIndexMap[playerId]
	if !exists {
		log.Printf("Warn - Tried to remove player %d, but they don't exist!", playerId)
		return
	}

	// Check if they are logged in
	player := &gamestate.players[playerIndex]
	if player.isLoggedIn() {
		// A little hacky, the setMenu() will trigger the world menu on exit
		player.setMenu(gamestate, PLAYER_MENU_LOGIN)
	}

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

func (gamestate *GameState) getPlayerByMobHandle(handle world.MobHandle) *Player {
	mob := gamestate.world.Mobs.Get(handle)
	if mob.PlayerCharacter == nil {
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

func (gamestate *GameState) messageRoom(roomIndex int, message string) {
	room := &gamestate.world.Rooms[roomIndex]

	for _, mobHandle := range room.Occupants {
		mob := gamestate.world.Mobs.Get(mobHandle)
		if mob.PlayerCharacter == nil {
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
