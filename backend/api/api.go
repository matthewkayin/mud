package api

import (
	"mud/game"
	"sync"
)

type ApiState struct {
	gamestate *game.GameState

	tokenToIdMutex sync.RWMutex
	tokenToIdMap map[string]int
}

func InitState(gamestate *game.GameState) *ApiState {
	return &ApiState {
		gamestate: gamestate,
		tokenToIdMutex: sync.RWMutex{},
		tokenToIdMap: make(map[string]int),
	}
}
