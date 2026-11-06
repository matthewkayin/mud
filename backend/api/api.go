package api

import (
	"sync"
	"mud/game"
)

type ApiState struct {
	gamestate *game.GameState
	sessionMutex sync.RWMutex
	// Session token to the account's username as stored in the account (its display form).
	// Sessions are only kept in memory, so restarting the server signs everyone out.
	sessionTokenToUsername map[string]string
}

func InitState(gamestate *game.GameState) *ApiState {
	return &ApiState {
		gamestate: gamestate,
		sessionTokenToUsername: make(map[string]string),
	}
}
