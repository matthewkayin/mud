package api

import (
	"mud/game"
	"sync"
)

type ApiState struct {
	gamestate *game.GameState

	// True when the server runs with --edit, which registers the editor's routes
	editorEnabled bool

	tokenToIdMutex sync.RWMutex
	tokenToIdMap map[string]int
}

func InitState(gamestate *game.GameState, editorEnabled bool) *ApiState {
	return &ApiState {
		gamestate: gamestate,
		editorEnabled: editorEnabled,
		tokenToIdMutex: sync.RWMutex{},
		tokenToIdMap: make(map[string]int),
	}
}
