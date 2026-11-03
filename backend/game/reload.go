package game

import (
	"log"
	"mud/world"
)

// Asks the game loop to reload the world. Sent by the editor's reload endpoint.
type ReloadRequest struct {
	// Receives the problems that stopped the reload, or none if it succeeded.
	// Must be buffered (size 1) so the game loop never blocks if the requester has gone away.
	Reply chan []string
}

// Rebuilds the world from the data folder. If the new world has problems, they are returned and the running
// world is untouched. Otherwise the new world replaces the old one, and every player in the world is
// spawned into it again from their save. Only transient state (NPCs, room items, combat, ...) is lost.
// Must only be called from the game loop.
func (gamestate *GameState) reloadWorld() []string {
	log.Printf("Reloading the world...")

	// Saving first means the new world loads online players' characters from their saves,
	// which also checks that they only reference content that still exists
	gamestate.saveAllLoggedInPlayers()

	newWorld, problems := world.WorldLoad()
	if len(problems) != 0 {
		log.Printf("Unable to reload the world because of these problems:")
		for _, problem := range problems {
			log.Print(problem)
		}
		return problems
	}

	gamestate.world.ScriptQuit()
	gamestate.world = newWorld

	for index := range len(gamestate.players) {
		player := &gamestate.players[index]
		switch player.menu {
			case PLAYER_MENU_WORLD: {
				gamestate.respawnPlayerAfterReload(player)
			}
			case PLAYER_MENU_CREATE: {
				// The character sheet holds race, class and job IDs, which aren't stable across reloads
				player.send("The world was reloaded, so your character creation was cancelled.")
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)
			}
		}
	}

	gamestate.broadcast("The world has been reloaded.")
	gamestate.handleWorldEvents()

	log.Printf("World reloaded.")
	return nil
}

// Spawns the player's character into the new world after a reload, in the room they were in.
// If that room no longer exists, they go to room 0.
func (gamestate *GameState) respawnPlayerAfterReload(player *Player) {
	// The player's trade partner is also in the world, and is reset by its own call
	player.tradeSession = nil
	player.isTalking = false

	// The old character belongs to the old world, so use the one the new world loaded from the save
	character, exists := gamestate.world.GetCharacterIfExists(player.character.Data.Name)
	if !exists {
		// This shouldn't happen, since the player was saved before the reload.
		// The menu is set directly because setMenu would run the world menu's onExit, which removes the old world's mob.
		log.Printf("Warn - Player %d's character %s has no save after the reload. Sending them to the login menu.",
			player.id, player.character.Data.Name)
		player.character = nil
		player.menu = PLAYER_MENU_LOGIN
		player.getMenu().onEnter(gamestate, player)
		return
	}
	player.character = character

	if character.Data.Room < 0 || character.Data.Room >= len(gamestate.world.Rooms) {
		character.Data.Room = 0
	}

	spawnPlayerMob(gamestate, player)
	describeRoomToPlayer(gamestate, player, &gamestate.world.Rooms[character.Data.Room])
}
