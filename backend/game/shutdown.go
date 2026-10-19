package game

import (
	"fmt"
	"log"
	"mud/world"
)

const SHUTDOWN_TICKS_PER_MINUTE = 60 / world.WORLD_SECONDS_PER_UPDATE

// Players are warned when this many minutes are left before a scheduled shutdown
var SHUTDOWN_WARNING_MINUTES = []int { 30, 15, 10, 5, 1 }

func (gamestate *GameState) scheduleShutdown(minutes int) {
	gamestate.serverMode = SERVER_MODE_SHUTDOWN_SCHEDULED
	gamestate.shutdownTimer = minutes * SHUTDOWN_TICKS_PER_MINUTE
	gamestate.broadcastShutdownWarning(minutes)
	log.Printf("Server shutdown scheduled in %d minute(s).", minutes)
}

func (gamestate *GameState) cancelShutdown() {
	gamestate.serverMode = SERVER_MODE_RUNNING
	gamestate.shutdownTimer = 0
	gamestate.broadcast("The scheduled server shutdown has been cancelled.")
	log.Printf("Server shutdown cancelled.")
}

func (gamestate *GameState) updateShutdown() {
	if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
		return
	}

	gamestate.shutdownTimer--
	if gamestate.shutdownTimer <= 0 {
		gamestate.enterMaintenance()
		return
	}

	for _, warningMinutes := range SHUTDOWN_WARNING_MINUTES {
		if gamestate.shutdownTimer == warningMinutes * SHUTDOWN_TICKS_PER_MINUTE {
			gamestate.broadcastShutdownWarning(warningMinutes)
		}
	}
}

func (gamestate *GameState) broadcastShutdownWarning(minutes int) {
	minutesUnit := "minutes"
	if minutes == 1 {
		minutesUnit = "minute"
	}

	gamestate.broadcast(fmt.Sprintf("The server will shut down for maintenance in %d %s.", minutes, minutesUnit))
}

// Saves and kicks every player, then refuses new connections until the server restarts
func (gamestate *GameState) enterMaintenance() {
	gamestate.serverMode = SERVER_MODE_MAINTENANCE
	gamestate.shutdownTimer = 0

	gamestate.broadcast("The server is shutting down for maintenance. Your character will be saved.")

	// Iterate backwards since removePlayer swaps the last player into the removed player's index
	for index := len(gamestate.players) - 1; index >= 0; index-- {
		player := &gamestate.players[index]
		gamestate.removePlayer(player.id, player.inbox)
	}

	log.Printf("Server entered maintenance mode.")
}
