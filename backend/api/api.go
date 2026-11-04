package api

type ApiState struct {
	// gamestate *game.GameState
	// true when server runs with --edit
	// editMode bool
	// tokenToIdMutex sync.RWMutex
	// tokenToIdMap map[string]int
}

// func InitState(gamestate *game.GameState, editMode bool) *ApiState {}
func InitState() *ApiState {
	return &ApiState {}
}
