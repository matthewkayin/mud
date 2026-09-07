package api

import (
	"mud/game"
	"sync"
)

type ApiState struct {
	gameState *game.GameState

	tokenToIdMutex sync.RWMutex
	tokenToIdMap map[string]int
}

func InitState(gameState *game.GameState) *ApiState {
	return &ApiState {
		gameState: gameState,
		tokenToIdMutex: sync.RWMutex{},
		tokenToIdMap: make(map[string]int),
	}
}
