package api

import (
	"sync"
)

type ApiState struct {
	// gamestate *game.GameState
	sessionMutex sync.RWMutex
	// Session token to the account's username as stored in the account (its display form).
	// Sessions are only kept in memory, so restarting the server signs everyone out.
	sessionTokenToUsername map[string]string
}

func InitState() *ApiState {
	return &ApiState {
		sessionTokenToUsername: make(map[string]string),
	}
}
