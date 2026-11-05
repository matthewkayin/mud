package world

import (
	"github.com/mmcdole/lunar"
)

const WORLD_SECONDS_PER_UPDATE = 3

type World struct {
	luaState *lua.State

	Characters map[string]*Character
	PlayerCharacters map[string][]string

	RaceData []*RaceData
	RaceNameToId map[string]RaceId

	ClassData []*ClassData
	ClassNameToId map[string]ClassId
}

func NewWorld() (*World, error) {
	world := &World{}

	world.Characters = make(map[string]*Character)
	world.PlayerCharacters = make(map[string][]string)

	// Init script state
	err := world.scriptInit()
	if err != nil {
		return nil, err
	}

	// Load script data
	loaders := []func() error {
		world.loadRaceData,
		world.loadClassData,
	}
	for _, loader := range loaders {
		err := loader()
		if err != nil {
			world.scriptQuit()
			return nil, err
		}
	}

	return world, nil
}

func (world *World) Quit() {
	world.scriptQuit()
}
